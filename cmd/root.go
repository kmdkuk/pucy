/*
Copyright © 2025 kmdkuk
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
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
	Run: func(cmd *cobra.Command, args []string) {
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		screen, err := tcell.NewScreen()
		if err != nil {
			panic(err)
		}
		if err := screen.Init(); err != nil {
			panic(err)
		}
		defer screen.Fini()

		keyword := ""
		selected := 0
		offset := 0 // scroll offset

		// Cache for filtered results by keyword
		filterCache := make(map[string][]string)

		// Helper to get filtered lines and their original indices
		getFiltered := func() []string {
			if filtered, ok := filterCache[keyword]; ok {
				return filtered
			}
			var filtered []string
			for _, line := range lines {
				if keyword == "" || strings.Contains(line, keyword) {
					filtered = append(filtered, line)
				}
			}
			filterCache[keyword] = filtered
			return filtered
		}

		draw := func() {
			screen.Clear()
			width, height := screen.Size()
			// 1 line for QUERY>
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
			putStr(screen, 0, 0, "QUERY> "+keyword)

			// Draw info at right top (add scroll info)
			scrollInfo := fmt.Sprintf("Total: %d  Filtered: %d  Scroll: %d/%d", len(lines), len(filtered), offset+1, len(filtered))
			putStr(screen, max(0, width-len(scrollInfo)-1), 0, scrollInfo) // -1 for margin
			y := 1
			for i := offset; i < len(filtered) && y < height; i++ {
				style := tcell.StyleDefault
				if i == selected {
					style = style.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
				}
				putStrStyled(screen, 0, y, filtered[i], style)
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
					return
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
					return
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

func putStr(s tcell.Screen, x, y int, str string) {
	screenWidth, screenHeight := s.Size()
	for i, r := range str {
		if x+i >= screenWidth || y >= screenHeight {
			break // Stop writing if we exceed screen boundaries
		}
		s.SetContent(x+i, y, r, nil, tcell.StyleDefault)
	}
}

// Add this helper function for styled output
func putStrStyled(s tcell.Screen, x, y int, str string, style tcell.Style) {
	screenWidth, screenHeight := s.Size()
	for i, r := range str {
		if x+i >= screenWidth || y >= screenHeight {
			break // Stop writing if we exceed screen boundaries
		}
		s.SetContent(x+i, y, r, nil, style)
	}
}
