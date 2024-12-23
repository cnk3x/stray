package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"slices"

	_ "embed"

	"github.com/cnk3x/stray/config"
	"github.com/cnk3x/stray/toast"

	"github.com/getlantern/systray"
	"github.com/samber/lo"
)

//go:generate go install -v github.com/akavel/rsrc@latest
//go:generate rsrc -ico .\icon.ico -manifest .\main.manifest -o .\main_windows.syso

var (
	//go:embed icon.ico
	icon []byte
	cfg  Shortcuts
)

func main() {
	var fn string
	flag.StringVar(&fn, "config", "", "config file")
	flag.Parse()

	if fn == "" {
		fn = "config.json"
	}

	if err := config.LoadFile(&cfg, fn); err != nil {
		Toast(err.Error(), "加载配置文件失败")
		return
	}

	fmt.Println("name", cfg.Name, cfg.Args["name"])

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	systray.Run(onReady(ctx), cancel)
}

func onReady(ctx context.Context) func() {
	return func() {
		systray.SetTitle(cfg.Name)
		systray.SetTooltip(cfg.Name)
		systray.SetIcon(icon)

		if len(cfg.Shortcuts) > 0 {
			systray.AddMenuItem(cfg.Name, cfg.Name).Disable()
			systray.AddSeparator()

			keys := lo.Keys(cfg.Shortcuts)
			slices.Sort(keys)

			for _, id := range keys {
				item := cfg.Shortcuts[id]
				AddMenuItem(ctx, item.Name, Action(func(menu *systray.MenuItem) {
					output, err := Run(ctx, item, cfg.Args)
					if err != nil {
						Toast(err.Error(), item.Name)
					} else {
						Toast(string(output), item.Name)
					}
				}))
			}
			systray.AddSeparator()
		}

		AddMenuItem(ctx, "退出", Action(func(*systray.MenuItem) { systray.Quit() }))
	}
}

var (
	re1 = regexp.MustCompile(` +`)
	re2 = regexp.MustCompile(`\n +`)
)

func Toast(msg string, title string) {
	n := toast.Notification{Title: title, Message: re2.ReplaceAllString(re1.ReplaceAllString(msg, " "), "\n")}
	err := n.Push()
	fmt.Println(msg)
	if err != nil {
		fmt.Println(err)
	}
}
