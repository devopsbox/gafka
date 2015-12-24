package command

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/funkygao/gafka/ctx"
	gzk "github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
	"github.com/samuel/go-zookeeper/zk"
)

type Ls struct {
	Ui  cli.Ui
	Cmd string

	zone      string
	path      string
	recursive bool
}

func (this *Ls) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("ls", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", "", "")
	cmdFlags.StringVar(&this.path, "p", "", "")
	cmdFlags.BoolVar(&this.recursive, "R", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if this.path == "" {
		this.path = "/"
	}

	if validateArgs(this, this.Ui).
		require("-z").
		invalid(args) {
		return 2
	}

	zkzone := gzk.NewZkZone(gzk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))
	defer zkzone.Close()
	conn := zkzone.Conn()

	if this.recursive {
		this.showChildrenRecursively(conn, this.path)
		return
	}

	children, _, err := conn.Children(this.path)
	must(err)
	sort.Strings(children)
	for _, child := range children {
		this.Ui.Output(child)
	}

	return
}

func (this *Ls) showChildrenRecursively(conn *zk.Conn, path string) {
	children, _, err := conn.Children(path)
	if err != nil {
		return
	}

	sort.Strings(children)
	for _, child := range children {
		if path == "/" {
			this.Ui.Output(path + child)
			this.showChildrenRecursively(conn, path+child)
		} else {
			this.Ui.Output(path + "/" + child)
			this.showChildrenRecursively(conn, path+"/"+child)
		}
	}
}

func (*Ls) Synopsis() string {
	return "List znode children"
}

func (this *Ls) Help() string {
	help := fmt.Sprintf(`
Usage: %s ls -z zone [options]

    List znode children

Options:

    -p path

    -R
      Recursive.    

`, this.Cmd)
	return strings.TrimSpace(help)
}