package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
		watchFlag       bool
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
	flag.BoolVar(&watchFlag, "watch", false, "Watch file for changes and automatically reload render")
	flag.BoolVar(&watchFlag, "W", false, "Watch file for changes (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: termstrap [options] <file.html|-|stdin>\n\n")
		fmt.Fprintf(os.Stderr, "Render HTML/CSS files or standard input directly in the terminal using Termstrap.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  termstrap file.html\n")
		fmt.Fprintf(os.Stderr, "  termstrap --watch file.html\n")
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

	buildModel := func(htmlContent string, w int) (*termstrap.Model, error) {
		m := &termstrap.Model{
			HTML:              htmlContent,
			Width:             w,
			Theme:             termstrap.Theme(themeFlag),
			RootPath:          rootPathFlag,
			DisableImages:     disableImgFlag,
			OptimizeSequences: &optimizeSeqFlag,
		}

		if cssFileFlag != "" {
			cssBytes, err := os.ReadFile(cssFileFlag)
			if err != nil {
				return nil, fmt.Errorf("reading CSS file %q: %w", cssFileFlag, err)
			}
			m.Stylesheets = append(m.Stylesheets, string(cssBytes))
		}

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
				m.ColorMode = cm
			}
		}
		imgOpts = append(imgOpts, termimage.WithOptimizeSequences(optimizeSeqFlag))
		if len(imgOpts) > 0 {
			m.ImageRenderer = termimage.NewRenderer(imgOpts...)
		}

		return m, nil
	}

	resolveWidth := func() int {
		if widthFlag > 0 {
			return widthFlag
		}
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			return w
		}
		return 80
	}

	if watchFlag {
		if inputFilePath == "" {
			fmt.Fprintf(os.Stderr, "Error: --watch requires an input file path, cannot watch stdin\n")
			os.Exit(1)
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		renderAndDisplay := func() (int, time.Time, int64, time.Time, int64, error) {
			curWidth := resolveWidth()

			fileInfo, err := os.Stat(inputFilePath)
			if err != nil {
				return curWidth, time.Time{}, 0, time.Time{}, 0, err
			}
			fMod := fileInfo.ModTime()
			fSize := fileInfo.Size()

			var cssMod time.Time
			var cssSize int64
			if cssFileFlag != "" {
				cssInfo, err := os.Stat(cssFileFlag)
				if err != nil {
					return curWidth, fMod, fSize, time.Time{}, 0, err
				}
				cssMod = cssInfo.ModTime()
				cssSize = cssInfo.Size()
			}

			contentBytes, err := os.ReadFile(inputFilePath)
			if err != nil {
				return curWidth, fMod, fSize, cssMod, cssSize, err
			}

			m, err := buildModel(string(contentBytes), curWidth)
			if err != nil {
				return curWidth, fMod, fSize, cssMod, cssSize, err
			}

			out, err := m.Render()
			if err != nil {
				return curWidth, fMod, fSize, cssMod, cssSize, err
			}

			// Clear screen and print
			fmt.Print("\033[H\033[2J")
			fmt.Print(out)

			return curWidth, fMod, fSize, cssMod, cssSize, nil
		}

		lastWidth, lastMod, lastSize, lastCSSMod, lastCSSSize, err := renderAndDisplay()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Initial render error: %v\n", err)
		}

		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-sigChan:
				fmt.Println()
				return
			case <-ticker.C:
				curW := resolveWidth()

				fi, err := os.Stat(inputFilePath)
				if err != nil {
					continue
				}

				changed := curW != lastWidth || fi.ModTime() != lastMod || fi.Size() != lastSize

				if !changed && cssFileFlag != "" {
					if cssFi, err := os.Stat(cssFileFlag); err == nil {
						if cssFi.ModTime() != lastCSSMod || cssFi.Size() != lastCSSSize {
							changed = true
						}
					}
				}

				if changed {
					time.Sleep(30 * time.Millisecond)
					newW, newMod, newSize, newCSSMod, newCSSSize, err := renderAndDisplay()
					if err == nil {
						lastWidth = newW
						lastMod = newMod
						lastSize = newSize
						lastCSSMod = newCSSMod
						lastCSSSize = newCSSSize
					}
				}
			}
		}
	}

	contentBytes, err := io.ReadAll(inputReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading content: %v\n", err)
		os.Exit(1)
	}

	width := resolveWidth()
	m, err := buildModel(string(contentBytes), width)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
