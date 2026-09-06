package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

func main() {
	var (
		widthFlag       int
		themeFlag       string
		rootPathFlag    string
		colorModeFlag   string
		protocolFlag    string
		cssFileFlag     string
		disableImgFlag  bool
		optimizeSeqFlag bool
	)

	flag.IntVar(&widthFlag, "width", 0, "Terminal width in columns (0 = auto-detect)")
	flag.IntVar(&widthFlag, "w", 0, "Terminal width in columns (shorthand)")
	flag.StringVar(&themeFlag, "theme", "bootstrap", "Visual theme (bootstrap, tokyonight, dracula)")
	flag.StringVar(&themeFlag, "t", "bootstrap", "Visual theme (shorthand)")
	flag.StringVar(&rootPathFlag, "root", "", "Root directory for resolving local relative image paths")
	flag.StringVar(&rootPathFlag, "r", "", "Root directory for resolving local images (shorthand)")
	flag.StringVar(&colorModeFlag, "color", "", "Color mode (truecolor, 256, 16)")
	flag.StringVar(&colorModeFlag, "c", "", "Color mode (shorthand)")
	flag.StringVar(&protocolFlag, "protocol", "", "Force image protocol (halfblock, kitty, iterm, sixel)")
	flag.StringVar(&protocolFlag, "p", "", "Force image protocol (shorthand)")
	flag.StringVar(&cssFileFlag, "css", "", "Path to custom CSS stylesheet file")
	flag.BoolVar(&disableImgFlag, "no-images", false, "Disable image rendering and replace with text placeholders")
	flag.BoolVar(&optimizeSeqFlag, "optimize", true, "Deduplicate contiguous ANSI escape sequences")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: termstrap [options] <file.html|-|stdin>\n\n")
		fmt.Fprintf(os.Stderr, "Render HTML/CSS files or standard input directly in the terminal using Termstrap.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  termstrap file.html\n")
		fmt.Fprintf(os.Stderr, "  termstrap -w 100 -t dracula file.html\n")
		fmt.Fprintf(os.Stderr, "  cat file.html | termstrap -\n")
		fmt.Fprintf(os.Stderr, "  make render FILE=file.html\n")
	}

	// Separate flags from positional arguments to support flags passed after filenames with spaces
	var flagsList []string
	var nonFlags []string
	rawArgs := os.Args[1:]
	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]
		if strings.HasPrefix(arg, "-") && arg != "-" {
			flagsList = append(flagsList, arg)
			if !strings.Contains(arg, "=") && i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				name := strings.TrimLeft(arg, "-")
				if name == "w" || name == "width" || name == "t" || name == "theme" ||
					name == "r" || name == "root" || name == "c" || name == "color" ||
					name == "p" || name == "protocol" || name == "css" {
					i++
					flagsList = append(flagsList, rawArgs[i])
				}
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}

	_ = flag.CommandLine.Parse(flagsList)

	// Determine input source
	var inputReader io.Reader
	var inputFilePath string

	if len(nonFlags) > 0 && nonFlags[0] != "-" {
		inputFilePath = nonFlags[0]
		f, err := os.Open(inputFilePath)
		if err != nil && os.IsNotExist(err) && len(nonFlags) > 1 {
			joined := strings.Join(nonFlags, " ")
			if fJoined, errJoined := os.Open(joined); errJoined == nil {
				f = fJoined
				err = nil
				inputFilePath = joined
			}
		}
		if err != nil && os.IsNotExist(err) && strings.Contains(inputFilePath, "%") {
			if unescaped, err2 := url.PathUnescape(inputFilePath); err2 == nil && unescaped != inputFilePath {
				if f2, err3 := os.Open(unescaped); err3 == nil {
					f = f2
					err = nil
					inputFilePath = unescaped
				}
			}
		}
		if err != nil && os.IsNotExist(err) && len(nonFlags) > 1 {
			joined := strings.Join(nonFlags, " ")
			if unescaped, err2 := url.PathUnescape(joined); err2 == nil {
				if f2, err3 := os.Open(unescaped); err3 == nil {
					f = f2
					err = nil
					inputFilePath = unescaped
				}
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %q: %v\n", inputFilePath, err)
			os.Exit(1)
		}
		defer f.Close()
		inputReader = f

		// Default root path to input file's directory if not explicitly provided
		if rootPathFlag == "" {
			rootPathFlag = filepath.Dir(inputFilePath)
		}
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 && (len(nonFlags) == 0 || nonFlags[0] != "-") {
			flag.Usage()
			os.Exit(1)
		}
		inputReader = os.Stdin
		if rootPathFlag == "" {
			rootPathFlag = "."
		}
	}

	contentBytes, err := io.ReadAll(inputReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading content: %v\n", err)
		os.Exit(1)
	}

	// Resolve terminal width
	width := widthFlag
	if width <= 0 {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			width = w
		} else {
			width = 80
		}
	}

	// Prepare Model
	model := termstrap.Model{
		HTML:              string(contentBytes),
		Width:             width,
		Theme:             termstrap.Theme(themeFlag),
		RootPath:          rootPathFlag,
		DisableImages:     disableImgFlag,
		OptimizeSequences: &optimizeSeqFlag,
	}

	// Custom CSS stylesheet
	if cssFileFlag != "" {
		cssBytes, err := os.ReadFile(cssFileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading CSS file %q: %v\n", cssFileFlag, err)
			os.Exit(1)
		}
		model.Stylesheets = append(model.Stylesheets, string(cssBytes))
	}

	// Image renderer options
	var imgOpts []termimage.Option
	if protocolFlag != "" {
		switch protocolFlag {
		case "halfblock":
			imgOpts = append(imgOpts, termimage.WithProtocol(termimage.HalfBlock))
		case "kitty":
			imgOpts = append(imgOpts, termimage.WithProtocol(termimage.Kitty))
		case "iterm", "iterm2":
			imgOpts = append(imgOpts, termimage.WithProtocol(termimage.ITerm2))
		case "sixel":
			imgOpts = append(imgOpts, termimage.WithProtocol(termimage.Sixel))
		}
	}
	if colorModeFlag != "" {
		if cm, ok := termimage.ParseColorMode(colorModeFlag); ok {
			imgOpts = append(imgOpts, termimage.WithColorMode(cm))
			model.ColorMode = cm
		}
	}
	imgOpts = append(imgOpts, termimage.WithOptimizeSequences(optimizeSeqFlag))

	if len(imgOpts) > 0 {
		model.ImageRenderer = termimage.NewRenderer(imgOpts...)
	}

	output, err := model.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
