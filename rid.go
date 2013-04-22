package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/holygeek/randomart"
	"github.com/holygeek/termsize"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func usage() {
	usage := `NAME
	rid - Show repository id (currently only grok git)

SYNOPSIS
	rid [-hr]

DESCRIPTION
	Show a repository unique id

OPTIONS
	-c
	  Clear screen before showing output

	-h
	  Show this help message

	-f
	  Flip output horizontally

	-r
	  Align output to the right

	-s <N>
	  Split sha1 sum into N-character strings. default is 10. Set to 0
	  to disable split
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
		die(err)
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

func die(args ...interface{}) {
	log.Fatal("rid", args)
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
			die("getReposig", err, "git", cmdsplit)
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
	chunkSize := flag.Int("s", 10, "Split sha1 sum into N-character strings")
	clearScreen := flag.Bool("c", false, "Clear screen before showing output")
	flip := flag.Bool("f", false, "Flip output horizontally")
	flag.Usage = usage
	flag.Parse()

	if *chunkSize == 0 {
		*chunkSize = 40
	}

	dirs := make([]string, 0)
	bytes, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	isGitRepo := err == nil
	if isGitRepo {
		gitDir := string(bytes)
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

	basename := mustGetBaseName(wd)
	chunks := splitSha1String(sha1str, *chunkSize)
	if *clearScreen {
		doClearScreen()
	}

	format := "%s\n"
	if *alignRight || *flip {
		ws := termsize.Get()
		format = fmt.Sprintf("%%%ds\n", ws.Col)
	}

	fmt.Printf(format, basename)
	for _, c := range chunks {
		fmt.Printf(format, mayReverse(*flip, c))
	}
	for _, str := range strings.Split(randomart, "\n") {
		fmt.Printf(format, mayReverse(*flip, str))
	}

}

func splitSha1String(sha1str string, chunkSize int) []string {
	sha1chunk := sha1str[:]
	nChunk := len(sha1str) / chunkSize
	chunks := make([]string, nChunk)
	for i := 0; i < nChunk; i++ {
		chunks[i] = sha1chunk[0:chunkSize]
		sha1chunk = sha1chunk[chunkSize:]
	}
	return chunks
}

func mayReverse(flip bool, str string) string {
	if !flip {
		return str
	}
	l := len(str)
	reversed := make([]byte, l)
	for i := 0; i < l; i++ {
		reversed[i] = str[l-i-1]
	}
	return string(reversed)
}

func doClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}
