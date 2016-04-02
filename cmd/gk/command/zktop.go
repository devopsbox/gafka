package command

import (
	"flag"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
	"github.com/funkygao/golib/color"
	"github.com/funkygao/termui"
	"github.com/nsf/termbox-go"
)

type Zktop struct {
	Ui  cli.Ui
	Cmd string

	refreshInterval time.Duration
	lastSents       map[string]string
	lastRecvs       map[string]string
}

func (this *Zktop) Run(args []string) (exitCode int) {
	var (
		zone  string
		graph bool
	)
	cmdFlags := flag.NewFlagSet("zktop", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", "", "")
	cmdFlags.DurationVar(&this.refreshInterval, "i", time.Second*5, "")
	cmdFlags.BoolVar(&graph, "g", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 2
	}

	this.lastRecvs = make(map[string]string)
	this.lastSents = make(map[string]string)

	if graph {
		var zkzones = make([]*zk.ZkZone, 0)
		if zone == "" {
			forSortedZones(func(zkzone *zk.ZkZone) {
				zkzones = append(zkzones, zkzone)
			})
		} else {
			zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
			zkzones = append(zkzones, zkzone)
		}

		this.draw(zkzones)
		return
	}

	for {
		refreshScreen()

		if zone == "" {
			forSortedZones(func(zkzone *zk.ZkZone) {
				this.displayZoneTop(zkzone)
			})
		} else {
			zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
			this.displayZoneTop(zkzone)
		}

		time.Sleep(this.refreshInterval)
	}

	return
}

func (this *Zktop) displayZoneTop(zkzone *zk.ZkZone) {
	this.Ui.Output(color.Green(zkzone.Name()))
	header := "VER             SERVER           PORT M  OUTST            RECVD             SENT CONNS  ZNODES LAT(MIN/AVG/MAX)"
	this.Ui.Output(header)

	stats := zkzone.RunZkFourLetterCommand("stat")
	sortedHosts := make([]string, 0, len(stats))
	for hp, _ := range stats {
		sortedHosts = append(sortedHosts, hp)
	}
	sort.Strings(sortedHosts)

	for _, hostPort := range sortedHosts {
		host, port, err := net.SplitHostPort(hostPort)
		if err != nil {
			panic(err)
		}

		stat := this.parsedStat(stats[hostPort])
		if stat.mode == "" {
			stat.mode = color.Red("E")
		} else if stat.mode == "L" {
			stat.mode = color.Blue(stat.mode)
		}
		var sentQps, recvQps int
		if lastRecv, present := this.lastRecvs[hostPort]; present {
			r1, _ := strconv.Atoi(stat.received)
			r0, _ := strconv.Atoi(lastRecv)
			recvQps = (r1 - r0) / int(this.refreshInterval.Seconds())

			s1, _ := strconv.Atoi(stat.sent)
			s0, _ := strconv.Atoi(this.lastSents[hostPort])
			sentQps = (s1 - s0) / int(this.refreshInterval.Seconds())
		}
		this.Ui.Output(fmt.Sprintf("%-15s %-15s %5s %1s %6s %16s %16s %5s %7s %s",
			stat.ver,                                     // 15
			host,                                         // 15
			port,                                         // 5
			stat.mode,                                    // 1
			stat.outstanding,                             // 6
			fmt.Sprintf("%s/%d", stat.received, recvQps), // 16
			fmt.Sprintf("%s/%d", stat.sent, sentQps),     // 16
			stat.connections,                             // 5
			stat.znodes,                                  // 7
			stat.latency,
		))

		this.lastRecvs[hostPort] = stat.received
		this.lastSents[hostPort] = stat.sent
	}
}

func (this *Zktop) draw(zkzones []*zk.ZkZone) {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	termui.UseTheme("helloworld")

	sinps := (func() []float64 {
		n := 220
		ps := make([]float64, n)
		for i := range ps {
			ps[i] = 1 + math.Sin(float64(i)/5)
		}
		return ps
	})()

	lc0 := termui.NewLineChart()
	lc0.Border.Label = "zk"
	lc0.Data = sinps
	lc0.Width = 50
	lc0.Height = 12
	lc0.X = 0
	lc0.Y = 0
	lc0.AxesColor = termui.ColorWhite
	lc0.LineColor = termui.ColorGreen | termui.AttrBold

	termui.Render(lc0)
	termbox.PollEvent()
}

type zkStat struct {
	ver            string
	latency        string
	connections    string
	outstanding    string
	mode           string
	znodes         string
	received, sent string
}

func (this *Zktop) parsedStat(s string) (stat zkStat) {
	lines := strings.Split(s, "\n")
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "Zookeeper version:"):
			p := strings.SplitN(l, ":", 2)
			p = strings.SplitN(p[1], ",", 2)
			stat.ver = strings.TrimSpace(p[0])

		case strings.HasPrefix(l, "Latency"):
			stat.latency = this.extractStatValue(l)

		case strings.HasPrefix(l, "Sent"):
			stat.sent = this.extractStatValue(l)

		case strings.HasPrefix(l, "Received"):
			stat.received = this.extractStatValue(l)

		case strings.HasPrefix(l, "Connections"):
			stat.connections = this.extractStatValue(l)

		case strings.HasPrefix(l, "Mode"):
			stat.mode = strings.ToUpper(this.extractStatValue(l)[:1])

		case strings.HasPrefix(l, "Node count"):
			stat.znodes = this.extractStatValue(l)

		case strings.HasPrefix(l, "Outstanding"):
			stat.outstanding = this.extractStatValue(l)

		}
	}
	return
}

func (this *Zktop) extractStatValue(l string) string {
	p := strings.SplitN(l, ":", 2)
	return strings.TrimSpace(p[1])
}

func (*Zktop) Synopsis() string {
	return "Unix “top” like utility for ZooKeeper"
}

func (this *Zktop) Help() string {
	help := fmt.Sprintf(`
Usage: %s zktop [options]

    Unix “top” like utility for ZooKeeper

Options:

    -z zone   

    -g
      Draws zk connections in graph. TODO

    -i interval
      Refresh interval in seconds.
      e,g. 5s

`, this.Cmd)
	return strings.TrimSpace(help)
}
