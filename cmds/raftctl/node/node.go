package node

import "github.com/spf13/cobra"

const (
	addrFlag = "addr"
	idFlag   = "id"
)

var (
	addr   string
	nodeID uint64

	NodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Request configuration changes for raftstore",
	}
)

func init() {
	NodeCmd.PersistentFlags().StringVar(&addr, addrFlag, "", "address of the DSS instance's raft admin endpoint, e.g. http://localhost:8082 (required)")
	NodeCmd.PersistentFlags().Uint64Var(&nodeID, idFlag, 0, "ID of the node to add/update/remove (required)")
	err := NodeCmd.MarkPersistentFlagRequired(addrFlag)
	cobra.CheckErr(err)
	err = NodeCmd.MarkPersistentFlagRequired(idFlag)
	cobra.CheckErr(err)

	NodeCmd.AddCommand(addCmd, updateCmd, removeCmd)
}
