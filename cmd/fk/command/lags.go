package command

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
	"github.com/funkygao/golib/gofmt"
	"github.com/ryanuber/columnize"
)

type Lags struct {
	Ui  cli.Ui
	Cmd string

	onlineOnly      bool
	groupPattern    string
	topicPattern    string
	problematicMode bool
	lagThreshold    int
}

func (this *Lags) Run(args []string) (exitCode int) {
	const secondsInMinute = 60
	var (
		cluster string
		zone    string
	)
	cmdFlags := flag.NewFlagSet("lags", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&cluster, "c", "", "")
	cmdFlags.BoolVar(&this.onlineOnly, "l", false, "")
	cmdFlags.BoolVar(&this.problematicMode, "p", false, "")
	cmdFlags.StringVar(&this.groupPattern, "g", "", "")
	cmdFlags.StringVar(&this.topicPattern, "t", "", "")
	cmdFlags.IntVar(&this.lagThreshold, "lag", 5000, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if validateArgs(this, this.Ui).
		require("-c", "-t").
		invalid(args) {
		return 2
	}

	if this.problematicMode {
		this.onlineOnly = true
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	zkcluster := zkzone.NewCluster(cluster) // panic if invalid cluster
	this.printConsumersLag(zkcluster)
	printSwallowedErrors(this.Ui, zkzone)

	return
}

func (this *Lags) printConsumersLag(zkcluster *zk.ZkCluster) {
	lines := make([]string, 0)
	header := "ConsumerGroup|Topic/Partition|Produced|Consumed|Lag|Refreshed"
	lines = append(lines, header)

	// sort by group name
	consumersByGroup := zkcluster.ConsumersByGroup(this.groupPattern)
	sortedGroups := make([]string, 0, len(consumersByGroup))
	for group, _ := range consumersByGroup {
		sortedGroups = append(sortedGroups, group)
	}
	sort.Strings(sortedGroups)

	for _, group := range sortedGroups {
		sortedTopicAndPartitionIds := make([]string, 0)
		consumers := make(map[string]zk.ConsumerMeta)
		for _, t := range consumersByGroup[group] {
			key := fmt.Sprintf("%s:%s", t.Topic, t.PartitionId)
			sortedTopicAndPartitionIds = append(sortedTopicAndPartitionIds, key)

			consumers[key] = t
		}
		sort.Strings(sortedTopicAndPartitionIds)

		for _, topicAndPartitionId := range sortedTopicAndPartitionIds {
			consumer := consumers[topicAndPartitionId]

			if !patternMatched(consumer.Topic, this.topicPattern) {
				continue
			}
			if !consumer.Online {
				continue
			}

			if consumer.Online {
				lines = append(lines,
					fmt.Sprintf("%s|%s/%s|%s|%s|%s|%s",
						group,
						consumer.Topic, consumer.PartitionId,
						gofmt.Comma(consumer.ProducerOffset),
						gofmt.Comma(consumer.ConsumerOffset),
						gofmt.Comma(consumer.Lag),
						gofmt.PrettySince(consumer.Mtime.Time()),
					))
			}
		}
	}

	this.Ui.Output(columnize.SimpleFormat(lines))
}

func (*Lags) Synopsis() string {
	return "Display online high level consumer group lags for a topic"
}

func (this *Lags) Help() string {
	help := fmt.Sprintf(`
Usage: %s lags [options]

    Display online high level consumer group lags for a topic

    e,g.
    %s lags -z prod -c trade -t OrderStatusMsg

Options:

    -z zone
      Default %s

    -c cluster

    -g consumer group name pattern

    -t topic name pattern

`, this.Cmd, this.Cmd, ctx.ZkDefaultZone())
	return strings.TrimSpace(help)
}