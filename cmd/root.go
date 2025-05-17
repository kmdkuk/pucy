/*
Copyright © 2025 kmdkuk
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kmdkuk/pucy/internal/matcher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pucy",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		screen, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		if err := screen.Init(); err != nil {
			return err
		}
		defer screen.Fini()

		keyword := ""
		selected := 0
		offset := 0 // scroll offset

		// Cache for filtered results by keyword
		filterCache := make(map[string][]string)
		m := matcher.NewMatcher()

		// Helper to get filtered lines and their original indices
		getFiltered := func() []string {
			if filtered, ok := filterCache[keyword]; ok {
				return filtered
			}
			var filtered []string
			linesMatchs := make(map[string]matcher.Matches, len(lines))
			for _, line := range lines {
				matchers := m.Match(line, keyword)
				if keyword == "" || len(matchers) > 0 {
					filtered = append(filtered, line)
					linesMatchs[line] = matchers
				}
			}
			sort.SliceStable(filtered, func(i, j int) bool {
				return linesMatchs[filtered[i]].MatchCount() > linesMatchs[filtered[j]].MatchCount()
			})
			filterCache[keyword] = filtered
			return filtered
		}

		draw := func() {
			screen.Clear()
			width, height := screen.Size()
			maxLines := height - 1
			filtered := getFiltered()

			// Adjust offset if selected is out of visible range
			if selected < offset {
				offset = selected
			}
			if selected >= offset+maxLines {
				offset = selected - maxLines + 1
			}
			if offset < 0 {
				offset = 0
			}

			// Draw search bar
			putStr(screen, 0, 0, "QUERY> "+keyword, tcell.StyleDefault)

			// Draw info at right top (add scroll info)
			scrollInfo := fmt.Sprintf("Total: %d  Filtered: %d  Scroll: %d/%d", len(lines), len(filtered), offset+1, len(filtered))
			putStr(screen, max(0, width-len(scrollInfo)-1), 0, scrollInfo, tcell.StyleDefault)
			y := 1
			for i := offset; i < len(filtered) && y < height; i++ {
				style := tcell.StyleDefault
				if i == selected {
					style = style.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
				}
				putStrHighlight(screen, 0, y, filtered[i], keyword, style)
				y++
			}
			screen.Show()
		}

		draw()
		for {
			ev := screen.PollEvent()
			switch tev := ev.(type) {
			case *tcell.EventKey:
				switch tev.Key() {
				case tcell.KeyEsc, tcell.KeyCtrlC:
					return nil
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					if len(keyword) > 0 {
						keyword = keyword[:len(keyword)-1]
						selected = 0
						offset = 0
						// Optionally, you can clear the cache here if you want to limit memory usage
					}
				case tcell.KeyEnter:
					filtered := getFiltered()
					if len(filtered) > 0 && selected < len(filtered) {
						screen.Fini() // Finish tcell screen before printing to stdout
						fmt.Println(filtered[selected])
					}
					return nil
				case tcell.KeyUp:
					if selected > 0 {
						selected--
					}
				case tcell.KeyDown:
					filtered := getFiltered()
					if selected < len(filtered)-1 {
						selected++
					}
				default:
					if tev.Rune() != 0 {
						keyword += string(tev.Rune())
						selected = 0
						offset = 0
						// Optionally, you can clear the cache here if you want to limit memory usage
					}
				}
				draw()
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pucy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".pucy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".pucy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func putStr(s tcell.Screen, x, y int, str string, style tcell.Style) {
	screenWidth, screenHeight := s.Size()
	for i, r := range str {
		if x+i >= screenWidth || y >= screenHeight {
			break // Stop writing if we exceed screen boundaries
		}
		s.SetContent(x+i, y, r, nil, style)
	}
}

// Helper function to print a string with keyword highlighting
func putStrHighlight(s tcell.Screen, x, y int, line, keyword string, style tcell.Style) {
	if keyword == "" {
		putStr(s, x, y, line, style)
		return
	}

	runes := []rune(line)
	matcher := matcher.NewMatcher()
	matches := matcher.Match(line, keyword)
	if len(matches) == 0 {
		return
	}
	i := 0
	pos := 0
	for pos < len(runes) {
		if matches.IsMatch(pos) {
			s.SetContent(x+i, y, runes[pos], nil, style.Foreground(tcell.ColorRed))
		} else {
			s.SetContent(x+i, y, runes[pos], nil, style)
		}
		i++
		pos++
	}
}

// Case-insensitive version of strings.Contains
func containsIgnoreCase(s, substr string) bool {
	// Convert both strings to lower case and use strings.Contains
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
