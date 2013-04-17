package main

import (
  "log"
  "fmt"
  "strings"
  "sort"
  "flag"
  "os"
  "os/exec"
  "crypto/sha1"
  "github.com/holygeek/randomart"
  "github.com/holygeek/termsize"
)

func usage() {
  usage := `NAME
  rid - Show repository id (currently only grok git)

SYNOPSIS
  rid [-hr]

DESCRIPTION
  Show a repository unique id

OPTIONS
  -h
    Show this help message

  -r
    Align output to the right
`
  fmt.Print(usage)
}

func mustSuccess(val interface{}, err error) interface{} {
  if err != nil {
    die(err)
  }
  return val
}

func getDirList(dirname string) []string {
  file, err := os.Open(dirname)
  if err != nil {
    die(err)
  }

  var fi os.FileInfo
  if fi, err = file.Stat(); err != nil {
    die(err)
  }

  if fi.IsDir() == false {
    log.Fatal("rid: not a directory: " + dirname)
  }

  dirs, err := file.Readdirnames(0)
  if err != nil {
    die(err);
  }

  ret := make([]string, 0)
  for _, dir := range dirs {
    fi, _ := os.Stat(dir)
    if fi.IsDir() {
      ret = append(ret, dir)
    }
  }

  return ret
}

func die (err error) {
  log.Fatal("rid: ", err)
}

func getRepoSig(dirs []string) string {
  logs := make([]string, len(dirs))
  for i, dir := range dirs {
    cmd := fmt.Sprintf("--git-dir=%s/.git --work-tree=%s log --no-decorate -1 --oneline",
                      dir, dir)

                      if *debug {
                        fmt.Println(cmd)
                      }
    cmdsplit := strings.Split(cmd, " ")
    out, err := exec.Command("git", cmdsplit...).Output()
    if err != nil {
      die(err)
    }
    logs[i] = string(out)
  }
  sort.Strings(logs)
  return strings.Join(logs, "\n")
}

func mustGetDirName(path string) string {
  lastSep := strings.LastIndex(path, "/")
  if lastSep == -1 {
    log.Fatal("rid: Could not get dirname from path '" + path + "'")
  }
  return path[0:lastSep]
}

func mustGetBaseName(path string) string {
  tokens := strings.Split(path, "/")
  if len(tokens) == 0 {
    log.Fatal("rid: Could not get basename from path '" + path + "'")
  }
  return tokens[len(tokens)-1]
}

var debug *bool
func main() {
  alignRight := flag.Bool("r", false, "Right align output")
  debug = flag.Bool("d", false, "Debug")
  flag.Usage = usage
  flag.Parse()

  dirs := make([]string, 0)
  bytes, err := exec.Command("git", "rev-parse", "--git-dir").Output()
  isGitRepo := err == nil
  if isGitRepo {
    gitDir := string(bytes);
    newLineIdx := strings.Index(gitDir, "\n")
    if newLineIdx != -1 {
      gitDir = strings.Split(gitDir, "\n")[0]
    }
    if gitDir == ".git" {
      dirs = append(dirs, ".")
    } else {
      dirs = append(dirs, mustGetDirName(gitDir))
    }
  } else {
    dirs = getDirList(".")
  }

  if *debug {
    fmt.Println("dir: '", dirs, "'")
  }
  reposig := getRepoSig(dirs)
  sha1 := sha1.New()
  sha1.Write([]byte(reposig))
  sha1str := fmt.Sprintf("%x", sha1.Sum(nil))
  randomart := randomart.FromString(reposig)

  wd, err := os.Getwd()
  if err != nil {
    die(err)
  }

  basename := mustGetBaseName(wd);
  if *alignRight {
    ws := termsize.Get()
    format := fmt.Sprintf("%%%ds\n", ws.Col)
    fmt.Printf(format, basename)
    fmt.Printf(format, sha1str)
    for _, str := range strings.Split(randomart, "\n") {
      fmt.Printf(format, str)
    }
  } else {
    fmt.Println(basename)
    fmt.Printf("%s\n", sha1str)
    fmt.Println(randomart)
  }
}
