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
	-C
	  Do not colorize first character in sha1 chunk

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

var (
	COLOR_BOLD_YELLOW = "\033[32;1m"
	COLOR_RESET       = "\033[0m"
)

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

		if opt.debug {
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

type Option struct {
	alignRight  bool
	debug       bool
	chunkSize   int
	noColor     bool
	clearScreen bool
	flip        bool
}

var opt = Option{chunkSize: 10}

const HEX_PER_CHAR = 2

func main() {
	flag.BoolVar(&opt.alignRight, "r", false, "Right align output")
	flag.BoolVar(&opt.debug, "d", false, "Debug")
	flag.IntVar(&opt.chunkSize, "s", opt.chunkSize, "Split sha1 sum into N-character strings")
	flag.BoolVar(&opt.noColor, "C", false, "Colorize first character in sha1 sum chunks")
	flag.BoolVar(&opt.clearScreen, "c", false, "Clear screen before showing output")
	flag.BoolVar(&opt.flip, "f", false, "Flip output horizontally")
	flag.Usage = usage
	flag.Parse()

	if opt.chunkSize == 0 {
		opt.chunkSize = sha1.Size * HEX_PER_CHAR
	}

	if opt.noColor {
		COLOR_BOLD_YELLOW = ""
		COLOR_RESET = ""
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

	if opt.debug {
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
	if opt.clearScreen {
		doClearScreen()
	}

	oneString, twoStrings := getFormatter(&opt)
	paint, reverse := getPainterAndReverser(&opt)

	fmt.Printf(oneString, basename)
	for _, c := range splitSha1String(sha1str, opt.chunkSize) {
		c = reverse(c)
		c = paint(c)
		fmt.Printf(twoStrings, "", c)
	}
	for _, str := range strings.Split(randomart, "\n") {
		str = reverse(str)
		fmt.Printf(oneString, str)
	}
}

func getFormatter(opt *Option) (oneString, twoStrings string) {
	oneString, twoStrings = "%s\n", "%s%s\n"
	if opt.alignRight || opt.flip {
		ws := termsize.Get()
		twoStrings = fmt.Sprintf("%%%ds%%s\n",
			ws.Col-uint16(opt.chunkSize))
		oneString = fmt.Sprintf("%%%ds\n", ws.Col)
	}
	return
}

func getPainterAndReverser(opt *Option) (paint, reverse func(string) string) {
	noop := func(str string) string { return str }
	paint, reverse = noop, noop
	if !opt.noColor {
		if opt.flip {
			paint = highlightLastChar
		} else {
			paint = highlightFirstChar
		}
	}
	if opt.flip {
		reverse = reverseString
	}
	return
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

func highlightFirstChar(str string) string {
	return COLOR_BOLD_YELLOW + str[0:1] + COLOR_RESET + str[1:]
}

func highlightLastChar(str string) string {
	l := len(str)
	return str[0:l-1] + COLOR_BOLD_YELLOW + str[l-1:] + COLOR_RESET
}

func reverseString(str string) string {
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
