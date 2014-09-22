package main

import (
    "os"
    "log"
    "time"
    "runtime"
    "os/signal"
    "github.com/Lupino/periodic/drivers"
    "github.com/Lupino/periodic"
    "github.com/Lupino/periodic/cmd/periodic/subcmd"
    "github.com/codegangsta/cli"
)


func main() {
    app := cli.NewApp()
    app.Name = "periodic"
    app.Usage = "Periodic task system"
    app.Version = "0.1.1"
    app.Flags = []cli.Flag {
        cli.StringFlag{
            Name: "H",
            Value: "unix:///tmp/periodic.sock",
            Usage: "the server address eg: tcp://127.0.0.1:5000",
            EnvVar: "PERIODIC_PORT",
        },
        cli.StringFlag{
            Name: "redis",
            Value: "tcp://127.0.0.1:6379",
            Usage: "The redis server address, required for driver redis",
        },
        cli.StringFlag{
            Name: "driver",
            Value: "leveldb",
            Usage: "The driver [leveldb, redis]",
        },
        cli.StringFlag{
            Name: "dbpath",
            Value: "leveldb",
            Usage: "The db path, required for driver leveldb",
        },
        cli.BoolFlag{
            Name: "d",
            Usage: "Enable daemon mode",
        },
        cli.IntFlag{
            Name: "timeout",
            Value: 0,
            Usage: "The socket timeout",
        },
        cli.IntFlag{
            Name: "cpus",
            Value: runtime.NumCPU(),
            Usage: "The runtime.GOMAXPROCS",
            EnvVar: "GOMAXPROCS",
        },
    }
    app.Commands = []cli.Command{
        {
            Name: "status",
            Usage: "Show status",
            Action: func(c *cli.Context) {
                subcmd.ShowStatus(c.GlobalString("H"))
            },
        },
        {
            Name: "submit",
            Usage: "Submit job",
            Flags: []cli.Flag {
                cli.StringFlag{
                    Name: "f",
                    Value: "",
                    Usage: "function name",
                },
                cli.StringFlag{
                    Name: "n",
                    Value: "",
                    Usage: "job name",
                },
                cli.StringFlag{
                    Name: "args",
                    Value: "",
                    Usage: "job workload",
                },
                cli.IntFlag{
                    Name: "t",
                    Value: 0,
                    Usage: "job running timeout",
                },
                cli.IntFlag{
                    Name: "periodiced_later",
                    Value: 0,
                    Usage: "job periodiced_later",
                },
            },
            Action: func(c *cli.Context) {
                var job = periodic.Job{
                    Name: c.String("n"),
                    Func: c.String("f"),
                    Args: c.String("args"),
                    Timeout: int64(c.Int("t")),
                }
                if len(job.Name) == 0 || len(job.Func) == 0 {
                    cli.ShowCommandHelp(c, "submit")
                    log.Fatal("Job name and func is require")
                }
                delay := c.Int("periodiced_later")
                var now = time.Now()
                job.SchedAt = int64(now.Unix()) + int64(delay)
                subcmd.SubmitJob(c.GlobalString("H"), job)
            },
        },
        {
            Name: "drop",
            Usage: "Drop func",
            Flags: []cli.Flag {
                cli.StringFlag{
                    Name: "f",
                    Value: "",
                    Usage: "function name",
                },
            },
            Action: func(c *cli.Context) {
                Func := c.String("f")
                if len(Func) == 0 {
                    cli.ShowCommandHelp(c, "drop")
                    log.Fatal("function name is required")
                }
                subcmd.DropFunc(c.GlobalString("H"), Func)
            },
        },
        {
            Name: "run",
            Usage: "Run func",
            Flags: []cli.Flag {
                cli.StringFlag{
                    Name: "f",
                    Value: "",
                    Usage: "function name required",
                },
                cli.StringFlag{
                    Name: "exec",
                    Value: "",
                    Usage: "command required",
                },
            },
            Action: func(c *cli.Context) {
                Func := c.String("f")
                exec := c.String("exec")
                if len(Func) == 0 {
                    cli.ShowCommandHelp(c, "run")
                    log.Fatal("function name is required")
                }
                if len(exec) == 0 {
                    cli.ShowCommandHelp(c, "run")
                    log.Fatal("command is required")
                }
                subcmd.Run(c.GlobalString("H"), Func, exec)
            },
        },
    }
    app.Action = func(c *cli.Context) {
        if c.Bool("d") {
            var st periodic.StoreDriver
            if c.String("driver") == "redis" {
                st = drivers.NewRedisDriver(c.String("redis"))
            } else {
                st = drivers.NewLevelDBDriver(c.String("dbpath"))
            }
            runtime.GOMAXPROCS(c.Int("cpus"))
            timeout := time.Duration(c.Int("timeout"))
            periodicd := periodic.NewSched(c.String("H"), st, timeout)
            go periodicd.Serve()
            s := make(chan os.Signal, 1)
            signal.Notify(s, os.Interrupt, os.Kill)
            <-s
            if c.String("driver") == "leveldb" {
                err := st.Close()
                if err != nil {
                    log.Fatal(err)
                }
            }
        } else {
            cli.ShowAppHelp(c)
        }
    }

    app.Run(os.Args)
}