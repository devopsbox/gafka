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

	zone        string
	path        string
	recursive   bool
	likePattern string
}

func (this *Ls) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("ls", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.BoolVar(&this.recursive, "R", false, "")
	cmdFlags.StringVar(&this.likePattern, "like", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(args) == 0 {
		this.Ui.Error("missing path")
		return 2
	}

	this.path = args[len(args)-1]

	if this.zone == "" {
		this.Ui.Error("unknown zone")
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
	if this.path == "/" {
		this.path = ""
	}
	for _, child := range children {
		this.Ui.Output(this.path + "/" + child)
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
			path = ""
		}

		znode := path + "/" + child

		if patternMatched(znode, this.likePattern) {
			this.Ui.Output(znode)
		}

		this.showChildrenRecursively(conn, znode)
	}
}

func (*Ls) Synopsis() string {
	return "List znode children"
}

func (this *Ls) Help() string {
	help := fmt.Sprintf(`
Usage: %s ls [options] path

    List znode children

Options:

    -z zone

    -R
      Recursively list subdirectories encountered.

    -like pattern
      Only display znode whose path is like this pattern.

`, this.Cmd)
	return strings.TrimSpace(help)
}
