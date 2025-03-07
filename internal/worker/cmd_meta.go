package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
)

type CmdType int

const (
	// Global represents a command that applies globally across all shards or nodes.
	// This type of command doesn't target a specific shard but affects the entire system.
	Global CmdType = iota

	// SingleShard represents a command that operates on a single shard.
	// This command is scoped to execute on one specific shard, optimizing for shard-local operations.
	SingleShard

	// MultiShard represents a command that operates across multiple shards.
	// This type of command spans more than one shard and may involve coordination between shards.
	MultiShard

	// Custom represents a command that is user-defined or has custom logic.
	// This command type allows for flexibility in executing specific, non-standard operations.
	Custom

	// Watch represents a command that is used to monitor changes or events.
	// This type of command listens for changes on specific keys or resources and responds accordingly.
	Watch

	// Unwatch represents a command that is used to stop monitoring changes or events.
	// This type of command stops listening for changes on specific keys or resources.
	Unwatch
)

// Global commands
const (
	CmdPing  = "PING"
	CmdAbort = "ABORT"
	CmdAuth  = "AUTH"
)

// Single-shard commands.
const (
	CmdExpire        = "EXPIRE"
	CmdExpireAt      = "EXPIREAT"
	CmdExpireTime    = "EXPIRETIME"
	CmdSet           = "SET"
	CmdGet           = "GET"
	CmdGetSet        = "GETSET"
	CmdGetEx         = "GETEX"
	CmdGetDel        = "GETDEL"
	CmdJSONArrAppend = "JSON.ARRAPPEND"
	CmdJSONArrLen    = "JSON.ARRLEN"
	CmdJSONArrPop    = "JSON.ARRPOP"
)

// Multi-shard commands.
const (
	CmdMset = "MSET"
	CmdMget = "MGET"
)

// Multi-Step-Multi-Shard commands
const (
	CmdRename = "RENAME"
	CmdCopy   = "COPY"
)

// Watch commands
const (
	CmdGetWatch      = "GET.WATCH"
	CmdGetUnWatch    = "GET.UNWATCH"
	CmdZRangeWatch   = "ZRANGE.WATCH"
	CmdHExists       = "HEXISTS"
	CmdHKeys         = "HKEYS"
	CmdHVals         = "HVALS"
	CmdZPopMin       = "ZPOPMIN"
	CmdJSONClear     = "JSON.CLEAR"
	CmdJSONStrlen    = "JSON.STRLEN"
	CmdJSONObjlen    = "JSON.OBJLEN"
	CmdZAdd          = "ZADD"
	CmdZRange        = "ZRANGE"
	CmdZRank         = "ZRANK"
	CmdZCount        = "ZCOUNT"
	CmdZRem          = "ZREM"
	CmdZCard         = "ZCARD"
	CmdPFAdd         = "PFADD"
	CmdPFCount       = "PFCOUNT"
	CmdPFMerge       = "PFMERGE"
	CmdTTL           = "TTL"
	CmdPTTL          = "PTTL"
	CmdIncr          = "INCR"
	CmdIncrBy        = "INCRBY"
	CmdDecr          = "DECR"
	CmdDecrBy        = "DECRBY"
	CmdIncrByFloat   = "INCRBYFLOAT"
	CmdHIncrBy       = "HINCRBY"
	CmdHIncrByFloat  = "HINCRBYFLOAT"
	CmdHRandField    = "HRANDFIELD"
	CmdGetRange      = "GETRANGE"
	CmdAppend        = "APPEND"
	CmdZPopMax       = "ZPOPMAX"
	CmdHLen          = "HLEN"
	CmdHStrLen       = "HSTRLEN"
	CmdHScan         = "HSCAN"
	CmdBFAdd         = "BF.ADD"
	CmdBFReserve     = "BF.RESERVE"
	CmdBFInfo        = "BF.INFO"
	CmdBFExists      = "BF.EXISTS"
	CmdCMSQuery      = "CMS.QUERY"
	CmdCMSInfo       = "CMS.INFO"
	CmdCMSInitByDim  = "CMS.INITBYDIM"
	CmdCMSInitByProb = "CMS.INITBYPROB"
	CmdCMSMerge      = "CMS.MERGE"
	CmdCMSIncrBy     = "CMS.INCRBY"
	CmdHSet          = "HSET"
	CmdHGet          = "HGET"
	CmdHSetnx        = "HSETNX"
	CmdHDel          = "HDEL"
	CmdHMSet         = "HMSET"
	CmdHMGet         = "HMGET"
)

type CmdMeta struct {
	CmdType
	Cmd                  string
	WorkerCommandHandler func([]string) []byte

	// decomposeCommand is a function that takes a DiceDB command and breaks it down into smaller,
	// manageable DiceDB commands for each shard processing. It returns a slice of DiceDB commands.
	decomposeCommand func(ctx context.Context, worker *BaseWorker, DiceDBCmd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error)

	// composeResponse is a function that combines multiple responses from the execution of commands
	// into a single response object. It accepts a variadic parameter of EvalResponse objects
	// and returns a unified response interface. It is used in the command type "MultiShard"
	composeResponse func(responses ...ops.StoreResponse) interface{}

	// preProcessingReq indicates whether the command requires preprocessing before execution.
	// If set to true, it signals that a preliminary step (such as fetching values from shards)
	// is necessary before the main command is executed. This is important for commands that depend
	// on the current state of data in the database.
	preProcessingReq bool

	// preProcessResponse is a function that handles the preprocessing of a DiceDB command by
	// preparing the necessary operations (e.g., fetching values from shards) before the command
	// is executed. It takes the worker and the original DiceDB command as parameters and
	// ensures that any required information is retrieved and processed in advance. Use this when set
	// preProcessingReq = true.
	preProcessResponse func(worker *BaseWorker, DiceDBCmd *cmd.DiceDBCmd)
}

var CommandsMeta = map[string]CmdMeta{
	// Global commands.
	CmdPing: {
		CmdType:              Global,
		WorkerCommandHandler: eval.RespPING,
	},

	// Single-shard commands.
	CmdSet: {
		CmdType: SingleShard,
	},
	CmdExpire: {
		CmdType: SingleShard,
	},
	CmdExpireAt: {
		CmdType: SingleShard,
	},
	CmdExpireTime: {
		CmdType: SingleShard,
	},
	CmdGet: {
		CmdType: SingleShard,
	},
	CmdGetSet: {
		CmdType: SingleShard,
	},
	CmdGetEx: {
		CmdType: SingleShard,
	},
	CmdGetDel: {
		CmdType: SingleShard,
	},
	CmdHExists: {
		CmdType: SingleShard,
	},
	CmdHKeys: {
		CmdType: SingleShard,
	},
	CmdHVals: {
		CmdType: SingleShard,
	},
	CmdJSONArrAppend: {
		CmdType: SingleShard,
	},
	CmdJSONArrLen: {
		CmdType: SingleShard,
	},
	CmdJSONArrPop: {
		CmdType: SingleShard,
	},
	CmdGetRange: {
		CmdType: SingleShard,
	},
	CmdJSONClear: {
		CmdType: SingleShard,
	},
	CmdJSONStrlen: {
		CmdType: SingleShard,
	},
	CmdJSONObjlen: {
		CmdType: SingleShard,
	},
	CmdPFAdd: {
		CmdType: SingleShard,
	},
	CmdPFCount: {
		CmdType: SingleShard,
	},
	CmdPFMerge: {
		CmdType: SingleShard,
	},
	CmdTTL: {
		CmdType: SingleShard,
	},
	CmdPTTL: {
		CmdType: SingleShard,
	},
	CmdHLen: {
		CmdType: SingleShard,
	},
	CmdHStrLen: {
		CmdType: SingleShard,
	},
	CmdHScan: {
		CmdType: SingleShard,
	},
	CmdHIncrBy: {
		CmdType: SingleShard,
	},
	CmdHIncrByFloat: {
		CmdType: SingleShard,
	},
	CmdHRandField: {
		CmdType: SingleShard,
	},

	// Multi-shard commands.
	CmdRename: {
		CmdType:            MultiShard,
		preProcessingReq:   true,
		preProcessResponse: preProcessRename,
		decomposeCommand:   decomposeRename,
		composeResponse:    composeRename,
	},

	CmdCopy: {
		CmdType:            MultiShard,
		preProcessingReq:   true,
		preProcessResponse: preProcessCopy,
		decomposeCommand:   decomposeCopy,
		composeResponse:    composeCopy,
	},

	CmdMset: {
		CmdType:          MultiShard,
		decomposeCommand: decomposeMSet,
		composeResponse:  composeMSet,
	},

	CmdMget: {
		CmdType:          MultiShard,
		decomposeCommand: decomposeMGet,
		composeResponse:  composeMGet,
	},
	CmdCMSQuery: {
		CmdType: SingleShard,
	},
	CmdCMSInfo: {
		CmdType: SingleShard,
	},
	CmdCMSIncrBy: {
		CmdType: SingleShard,
	},
	CmdCMSInitByDim: {
		CmdType: SingleShard,
	},
	CmdCMSInitByProb: {
		CmdType: SingleShard,
	},
	CmdCMSMerge: {
		CmdType: SingleShard,
	},
	CmdHSet: {
		CmdType: SingleShard,
	},
	CmdHGet: {
		CmdType: SingleShard,
	},
	CmdHSetnx: {
		CmdType: SingleShard,
	},
	CmdHDel: {
		CmdType: SingleShard,
	},
	CmdHMSet: {
		CmdType: SingleShard,
	},
	CmdHMGet: {
		CmdType: SingleShard,
	},

	// Custom commands.
	CmdAbort: {
		CmdType: Custom,
	},
	CmdAuth: {
		CmdType: Custom,
	},

	// Watch commands
	CmdGetWatch: {
		CmdType: Watch,
	},
	CmdZRangeWatch: {
		CmdType: Watch,
	},

	// Unwatch commands
	CmdGetUnWatch: {
		CmdType: Unwatch,
	},

	// Sorted set commands
	CmdZAdd: {
		CmdType: SingleShard,
	},
	CmdZCount: {
		CmdType: SingleShard,
	},
	CmdZRank: {
		CmdType: SingleShard,
	},
	CmdZRange: {
		CmdType: SingleShard,
	},
	CmdZCard: {
		CmdType: SingleShard,
	},
	CmdZRem: {
		CmdType: SingleShard,
	},
	CmdAppend: {
		CmdType: SingleShard,
	},
	CmdIncr: {
		CmdType: SingleShard,
	},
	CmdIncrBy: {
		CmdType: SingleShard,
	},
	CmdDecr: {
		CmdType: SingleShard,
	},
	CmdDecrBy: {
		CmdType: SingleShard,
	},
	CmdIncrByFloat: {
		CmdType: SingleShard,
	},
	CmdZPopMin: {
		CmdType: SingleShard,
	},
	CmdZPopMax: {
		CmdType: SingleShard,
	},

	// Bloom Filter
	CmdBFAdd: {
		CmdType: SingleShard,
	},
	CmdBFInfo: {
		CmdType: SingleShard,
	},
	CmdBFExists: {
		CmdType: SingleShard,
	},
	CmdBFReserve: {
		CmdType: SingleShard,
	},
}

func init() {
	for c, meta := range CommandsMeta {
		if err := validateCmdMeta(c, meta); err != nil {
			slog.Error("error validating worker command metadata %s: %v", c, err)
		}
	}
}

// validateCmdMeta ensures that the metadata for each command is properly configured
func validateCmdMeta(c string, meta CmdMeta) error {
	switch meta.CmdType {
	case Global:
		if meta.WorkerCommandHandler == nil {
			return fmt.Errorf("global command %s must have WorkerCommandHandler function", c)
		}
	case MultiShard:
		if meta.decomposeCommand == nil || meta.composeResponse == nil {
			return fmt.Errorf("multi-shard command %s must have both decomposeCommand and composeResponse implemented", c)
		}
	case SingleShard, Watch, Unwatch, Custom:
		// No specific validations for these types currently
	default:
		return fmt.Errorf("unknown command type for %s", c)
	}

	return nil
}
