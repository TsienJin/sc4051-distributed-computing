package monitor

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log/slog"
	"os"
	"server/internal/bookings"
	"server/internal/vars"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mu sync.RWMutex // To ensure that commands are executed mutually.

	envEnableDuplicateFiltering  bool
	envDisableDuplicateFiltering bool
	envPacketDropRate            float32
	envPacketReceiveTimeout      int
	envPacketTTL                 int
	envMessageAssemblerIntervals int
	envResponseTTL               int
	envResponseIntervals         int

	flagEnableDuplicateFiltering  string = "enable-duplicate-filtering"
	flagDisableDuplicateFiltering string = "disable-duplicate-filtering"
	flagPacketDropRate            string = "packet-drop-rate"
	flagPacketReceiveTimeout      string = "packet-receive-timeout"
	flagPacketTTL                 string = "packet-ttl"
	flagMessageAssemblerIntervals string = "message-assembler-intervals"
	flagResponseTTL               string = "response-ttl"
	flagResponseIntervals         string = "response-intervals"
)

var (
	tableHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).PaddingLeft(1).PaddingRight(1).Bold(true).Align(lipgloss.Left)
	tableCellStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1).Align(lipgloss.Left)
)

func newTable() *table.Table {
	return table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return tableHeaderStyle
			default:
				return tableCellStyle
			}
		})

}

func init() {
	register()
}

func register() {
	// Register command hierarchy
	rootCmd.SetHelpCommand(helpCmd)
	rootCmd.AddCommand(envRootCmd, recordsCmd, resetRootCmd, nukeRootCmd, networkCmd)

	// Add subcommands for env
	envRootCmd.AddCommand(envShowCmd, envSetCmd)
	envSetCmd.Flags().BoolVar(&envEnableDuplicateFiltering, flagEnableDuplicateFiltering, true, "Enable duplicate packet/message filtering")
	envSetCmd.Flags().BoolVar(&envDisableDuplicateFiltering, flagDisableDuplicateFiltering, false, "Disable duplicate packet/message filtering")
	envSetCmd.Flags().Float32Var(&envPacketDropRate, flagPacketDropRate, 0.0, "Set the network drop rate value")
	envSetCmd.Flags().IntVar(&envPacketReceiveTimeout, flagPacketReceiveTimeout, 0, "Set packet receive timeout (ms)")
	envSetCmd.Flags().IntVar(&envPacketTTL, flagPacketTTL, 0, "Set packet TTL (ms)")
	envSetCmd.Flags().IntVar(&envMessageAssemblerIntervals, flagMessageAssemblerIntervals, 0, "Set message assembler intervals (ms)")
	envSetCmd.Flags().IntVar(&envResponseTTL, flagResponseTTL, 0, "Set response TTL (ms)")
	envSetCmd.Flags().IntVar(&envResponseIntervals, flagResponseIntervals, 0, "Set response intervals (ms)")

	// Add subcommands for reset
	resetRootCmd.AddCommand(resetAllCmd, resetRecordsCmd, resetNetCmd)
}

func ExecuteUserCommand(line string) string {
	mu.Lock()
	defer mu.Unlock()

	// reset flags to make sure next execution is "fresh"
	// TODO: subcommand help flag is not reset. DO NOT USE --help in subcommands
	defer func() {
		reset := func(cmd *cobra.Command) {
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				f.Changed = false
			})
			// If you have persistent flags you need to reset them as well:
			cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				f.Changed = false
			})
		}

		for _, c := range []*cobra.Command{
			rootCmd,
			helpCmd,
			envRootCmd,
			envShowCmd,
			envSetCmd,
			resetRootCmd,
			resetAllCmd,
			resetRecordsCmd,
			resetNetCmd,
			nukeRootCmd,
		} {
			reset(c)
		}

	}()

	// Ensure that line has leading '/'
	if (len(line) > 0 && line[0] != '/') || line == "" {
		slog.Warn("Attempted to parse malformed user command as command")
		return ""
	}

	// Strip prefix '/'
	line = strings.TrimPrefix(line, "/")

	// Parse line into args
	parser := shellwords.NewParser()
	args, err := parser.Parse(line)
	if err != nil {
		slog.Error("Unable to parse command", "line", line, "err", err)
		return ""
	}

	// Set output buffer and execute args
	buf := bytes.Buffer{}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Unable to execute command", "line", line, "err", err)
		return ""
	}

	return buf.String()
}

var rootCmd = &cobra.Command{
	Use:   "/",
	Short: "Interact with the server over TCP",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			return
		}
	},
}

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help menu",
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.Help()
		if err != nil {
			return
		}
	},
}

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Show all current facility and booking records in memory",
	Run: func(cmd *cobra.Command, args []string) {

		singaporeTimeZone := time.FixedZone("UTC+8", 8*60*60)

		facilitiesTable := newTable().Headers("NAME", "NO. BOOKINGS")
		bookingTable := newTable().Headers("FACILITY", "BOOKING ID", "START", "END")

		records := bookings.GetManager().GetDeepCopyOfRecords()

		for fName, f := range records {
			facilitiesTable = facilitiesTable.Row(string(fName), fmt.Sprintf("%v", len(f.Bookings)))

			for _, b := range f.Bookings {
				bookingTable = bookingTable.Row(
					string(fName),
					strconv.Itoa(int(b.Id)),
					b.Start.In(singaporeTimeZone).Format("2006-01-02 15:04:05"),
					b.End.In(singaporeTimeZone).Format("2006-01-02 15:04:05"),
				)
			}
		}

		// return table
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), facilitiesTable.String()+"\n")
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), bookingTable.String())

	},
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Show network statistics",
	Run: func(cmd *cobra.Command, args []string) {
		stats := getNetworkStats()

		var inDropPercentage float64 = 0
		var outDropPercentage float64 = 0

		if stats.packetInExpected != 0 {
			inDropPercentage = 100 * (float64(stats.packetInDropped) / (float64(stats.packetInExpected)))
		}
		if stats.packetOutExpected != 0 {
			outDropPercentage = 100 * (float64(stats.packetOutDropped) / (float64(stats.packetOutExpected)))
		}

		t := newTable().
			Headers("DIRECTION", "EXPECTED", "DROPPED").
			Row(
				"IN",
				strconv.Itoa(stats.packetInExpected),
				fmt.Sprintf("%d\t(%.2f PERCENT)", stats.packetInDropped, inDropPercentage),
			).
			Row(
				"OUT",
				strconv.Itoa(stats.packetOutExpected),
				fmt.Sprintf("%d\t(%.2f PERCENT)", stats.packetOutDropped, outDropPercentage),
			)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.String())
	},
}

var envRootCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage server environment settings",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			return
		}
	},
}

var envShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Prints current environment configs for the server",
	Run: func(cmd *cobra.Command, args []string) {

		envVars := vars.GetStaticEnvCopy()

		t := newTable()
		t = t.Headers("ENV VAR", "VALUE")
		t = t.Rows([][]string{
			{"EnableDuplicateFiltering", fmt.Sprintf("%v", envVars.EnableDuplicateFiltering)},
			{"PacketDropRate", fmt.Sprintf("%v", envVars.PacketDropRate)},
			{"PacketReceiveTimeout", fmt.Sprintf("%v", envVars.PacketReceiveTimeout)},
			{"PacketTTL", fmt.Sprintf("%v", envVars.PacketTTL)},
			{"MessageAssemblerIntervals", fmt.Sprintf("%v", envVars.MessageAssemblerIntervals)},
			{"ResponseTTL", fmt.Sprintf("%v", envVars.ResponseTTL)},
			{"ResponseIntervals", fmt.Sprintf("%v", envVars.ResponseIntervals)},
		}...)

		_, err := fmt.Fprintf(cmd.OutOrStdout(), t.String())
		if err != nil {
			return
		}

	},
}

var envSetCmd = &cobra.Command{
	Use:   "set [flags]",
	Short: "Sets configs for the server",
	Run: func(cmd *cobra.Command, args []string) {

		sendErrToBuffer := func(err error) {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), err.Error())
			if err != nil {
				return
			}
		}

		cmd.Flags().Visit(func(f *pflag.Flag) {
			if !f.Changed {
				return
			}
			switch f.Name {
			case "enable-duplicate-filtering":
				err := vars.SetEnableDuplicateFiltering(envEnableDuplicateFiltering)
				if err != nil {
					sendErrToBuffer(err)
				}
			case "disable-duplicate-filtering":
				err := vars.SetEnableDuplicateFiltering(!envDisableDuplicateFiltering)
				if err != nil {
					sendErrToBuffer(err)
				}
			case "packet-drop-rate":
				floatVal, err := strconv.ParseFloat(f.Value.String(), 32)
				if err != nil {
					sendErrToBuffer(err)
				}
				err = vars.SetPacketDropRate(float32(floatVal))
				if err != nil {
					sendErrToBuffer(err)
				}
			case "packet-receive-timeout":
				val, err := strconv.Atoi(f.Value.String())
				if err != nil {
					sendErrToBuffer(err)
				}
				if err := vars.SetPacketReceiveTimeout(val); err != nil {
					sendErrToBuffer(err)
				}
			case "packet-ttl":
				val, err := strconv.Atoi(f.Value.String())
				if err != nil {
					sendErrToBuffer(err)
				}
				if err := vars.SetPacketTTL(val); err != nil {
					sendErrToBuffer(err)
				}
			case "message-assembler-intervals":
				val, err := strconv.Atoi(f.Value.String())
				if err != nil {
					sendErrToBuffer(err)
				}
				if err := vars.SetMessageAssemblerIntervals(val); err != nil {
					sendErrToBuffer(err)
				}
			case "response-ttl":
				val, err := strconv.Atoi(f.Value.String())
				if err != nil {
					sendErrToBuffer(err)
				}
				if err := vars.SetResponseTTL(val); err != nil {
					sendErrToBuffer(err)
				}
			case "response-intervals":
				val, err := strconv.Atoi(f.Value.String())
				if err != nil {
					sendErrToBuffer(err)
				}
				if err := vars.SetResponseIntervals(val); err != nil {
					sendErrToBuffer(err)
				}
			default:
				sendErrToBuffer(fmt.Errorf("%s flag not supposed by envSetCmd", f.Name))
			}
		})
	},
}

var resetRootCmd = &cobra.Command{
	Use:   "reset [domain]",
	Short: "Resets the specified domain",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			return
		}
	},
}

var resetAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Resets bookings, facilities, and network stats",
	Run: func(cmd *cobra.Command, args []string) {
		bookings.GetMonitor().Reset()
		resetNetworkMonitor()
	},
}

var resetRecordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Resets bookings and facilities",
	Run: func(cmd *cobra.Command, args []string) {
		bookings.GetMonitor().Reset()
	},
}

var resetNetCmd = &cobra.Command{
	Use:   "net",
	Short: "Resets network stats",
	Run: func(cmd *cobra.Command, args []string) {
		resetNetworkMonitor()
	},
}

var nukeRootCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Triggers the server to exit (will restart when using restart policy in Docker)",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Warn("THIS WAS AN EXPECTED OS.EXIT(1), TRIGGERED FROM `handle_client_cmd.nukeRootCmd`")
		os.Exit(1)
	},
}
