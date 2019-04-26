package main

// imports //////////////////////////////////////

import (
  "os"
  "fmt"
  "bufio"
  "strings"
  "crypto/md5"

  "github.com/spf13/cobra"
  "github.com/Shopify/sysv_mq"
  "github.com/siadat/ipc"

  "github.com/callowaylc/mq/pkg"
  "github.com/callowaylc/mq/pkg/log"
)

// constants ////////////////////////////////////

const ExitStatusArgument int = 30
const ExitStatusArgumentStdin int = 31
const ExitStatusSysv int = 40
const ExitStatusSysvRead int = 41
const ExitStatusSysvWrite int = 42
const ExitStatusSysvStat int = 43
const ExitStatusSysvDelete int = 44
const ExitStatusSysvCount int = 45


const LeastSignificantId uint64 = 42


// main /////////////////////////////////////////

func init() {
  log.Init()
}

func main() {
  logger := log.Logger(pkg.Trace("main.main", "main"))
  logger.Info().Msg("Enter")
  defer logger.Info().Msg("Exit")

  var fcount bool
  var fdelete bool
  var fdump bool

  root := &cobra.Command{
    Use: "mq key [value]",

    // define logger behavior, which is to parse message and
    // write to stderr
    Run: func(cmd *cobra.Command, args []string) {
      logger := log.Logger(pkg.Trace("main.Run", "main"))
      logger.Info().
        Str("args", fmt.Sprint(args)).
        Msg("Enter")
      defer logger.Info().Msg("Exit")

      if len(args) >= 1 {
        key, err := determineKey(args[0])
        if err != nil {
          logger.Error().
            Str("name", args[0]).
            Int("key", int(key)).
            Str("error", err.Error()).
            Msg("Failed to generate key")
          os.Exit(ExitStatusArgument)
        }
        logger.Info().
          Str("name", args[0]).
          Int("key", int(key)).
          Msg("Created IPC key")

        // attempt to create queue
        mq, err := sysv_mq.NewMessageQueue(&sysv_mq.QueueConfig{
          Key:     int(key),
          MaxSize: 1024,
          Mode:    sysv_mq.IPC_CREAT | 0660,
        })
        if err != nil {
          logger.Error().
            Str("name", args[0]).
            Int("key", int(key)).
            Str("error", err.Error()).
            Msg("Failed to determine message queue")
          os.Exit(ExitStatusSysv)
        }
        logger.Info().
          Str("name", args[0]).
          Int("key", int(key)).
          Msg("Determined message queue")


        // otherwise, determine if read or write operation
        // against queue
        messages  := []string{}
        stat, err := os.Stdin.Stat()
        if err != nil {
          logger.Error().
            Str("name", args[0]).
            Int("key", int(key)).
            Str("mechanism", "stdin").
            Msg("Failed to stat stdin")
        }

        //if (stat.Mode() & os.ModeCharDevice) == 0 {
        if err == nil && stat.Size() > 0 {
          // NOTE: when tty isnt attached,
          logger.Info().
            Str("name", args[0]).
            Int("key", int(key)).
            Str("mechanism", "stdin").
            Int("size", int(stat.Size())).
            Msg("Preparing write operation")

          in := bufio.NewScanner(os.Stdin)
          for in.Scan() {
            message := strings.TrimSpace(in.Text())
            messages = append(messages, message)
            logger.Info().
              Str("name", args[0]).
              Int("key", int(key)).
              Str("line", message).
              Msg("Read from STDIN")
          }
          if in.Err() != nil {
            logger.Info().
              Str("name", args[0]).
              Int("key", int(key)).
              Str("error", fmt.Sprint(in.Err())).
              Msg("Encountered an error while reading from STDIN")
            os.Exit(ExitStatusArgumentStdin)
          }

        } else if len(args) == 2 {
          logger.Info().
            Str("name", args[0]).
            Int("key", int(key)).
            Str("mechanism", "argument").
            Msg("Preparing write operation")
          messages = append(messages, args[1])
        }

        if len(messages) > 0 {
          logger.Info().
            Str("name", args[0]).
            Int("key", int(key)).
            Int("count", len(messages)).
            Msg("Performing write operation")

          // we are performing a write operations against
          // message slice
          for _, m := range messages {
            logger.Debug().
              Str("name", args[0]).
              Int("key", int(key)).
              Str("payload", m).
              Msg("Write message to queue")

            err := mq.SendString(m, 1, 0)
            if err != nil {
              logger.Error().
                Str("name", args[0]).
                Int("key", int(key)).
                Str("payload", m).
                Msg("Failed to send message")
              os.Exit(ExitStatusSysvWrite)
            }
          }

        } else if len(args) == 1 {

          // otherwise we are performing a read
          // operation
          logger.Info().
            Str("name", args[0]).
            Int("key", int(key)).
            Msg("Performing read operation")

          if fdelete {
            logger.Info().
              Str("name", args[0]).
              Int("key", int(key)).
              Msg("Signal queue deletion")

            // if delete flag has been specified, defer deltion
            defer func() {
              err := mq.Destroy()
              if err != nil {
                logger.Error().
                  Str("name", args[0]).
                  Int("key", int(key)).
                  Str("error", err.Error()).
                  Msg("Failed to delete queue")
                os.Exit(ExitStatusSysvDelete)
              }
              logger.Info().
                Str("name", args[0]).
                Int("key", int(key)).
                Msg("Queue deleted")

              path := seedFile(args[0])
              err = os.Remove(path)
              if err != nil {
                logger.Error().
                  Str("name", args[0]).
                  Int("key", int(key)).
                  Str("error", err.Error()).
                  Msg("Failed to delete seed file")
                os.Exit(ExitStatusSysvDelete)
              }
              logger.Info().
                Str("name", args[0]).
                Int("key", int(key)).
                Str("path", path).
                Msg("Deleted seed file")

            }()

          } else if fcount {
            // Performing queue count
            logger.Info().
              Str("name", args[0]).
              Int("key", int(key)).
              Msg("Performing count operation")

            count, err := mq.Count()
            if err != nil {
              logger.Error().
                Str("name", args[0]).
                Int("key", int(key)).
                Str("error", err.Error()).
                Msg("Failed to get queue count")
              os.Exit(ExitStatusSysvCount)
            }

            // write count to stdout
            fmt.Println(count)

          } else {
            // Performing queue count
            logger.Info().
              Str("name", args[0]).
              Int("key", int(key)).
              Msg("Performing dequeue operation")

            var count uint64 = 1
            if fdump {
              // if dump has been passed we dump all messages
              // to stdout and delete queue
              logger.Info().
                Str("name", args[0]).
                Int("key", int(key)).
                Msg("Dumping queue")

              count, err = mq.Count()
              if err != nil {
                logger.Error().
                  Str("name", args[0]).
                  Int("key", int(key)).
                  Str("error", err.Error()).
                  Msg("Failed to get queue count")
                os.Exit(ExitStatusSysvCount)
              }
            }
            logger.Info().
              Str("name", args[0]).
              Int("key", int(key)).
              Int("size", int(count)).
              Msg("Dequeue count")

            for counter := 0; counter < int(count); counter++ {
              message, mtype, err := mq.ReceiveString(1, 0)
              if err != nil {
                logger.Error().
                  Str("name", args[0]).
                  Int("key", int(key)).
                  Int("type", mtype).
                  Str("error", err.Error()).
                  Int("size", int(count)).
                  Msg("Failed to read message")
                os.Exit(ExitStatusSysvRead)
              }
              logger.Debug().
                Str("name", args[0]).
                Int("key", int(key)).
                Int("size", int(count)).
                Int("counter", counter).
                Str("payload", message).
                Msg("Dequeue message")

              // write message to stdout
              fmt.Println(message)
            }
          }

        } else {
          // log wasted arguments - this can be useful when
          // interacting with the binary from a shell environment
          logger.Error().
            Str("name", args[0]).
            Int("key", int(key)).
            Str("args", fmt.Sprint(args)).
            Msg("Unexpected arguments")
          os.Exit(ExitStatusArgument)
        }

      } else {
        // if no stdin and no arguments then we display our help message
        // and exit with a failed status code
        cmd.Help()
        os.Exit(ExitStatusArgument)
      }
    },
  }

  root.PersistentFlags().BoolVarP(
    &fcount, "count", "c", false, "Number of messages in the queue",
  )
  root.PersistentFlags().BoolVarP(
    &fdelete, "delete", "d", false, "Delete the queue",
  )
  root.PersistentFlags().BoolVarP(
    &fdump, "dump", "", false, "Dump and delete the queue",
  )
  root.Execute()
}

func determineKey(name string) (uint64, error) {
  // create sysv ipc key
  logger := log.Logger(pkg.Trace("main.determineKey", "main"))
  logger.Info().
    Str("name", name).
    Msg("Enter")
  defer logger.Info().Msg("Exit")

  path := seedFile(name)
  seed, err := os.Create(path)
  if err != nil {
    return 0, err
  }
  defer func() {
    seed.Close()
  }()

  logger.Info().
    Str("path", path).
    Msg("Create seed file")

  // https://github.com/siadat/ipc/blob/master/ftok.go
  return ipc.Ftok(path, LeastSignificantId)
}

func seedFile(name string) string {
  // a seed file required for generating an mq id
  logger := log.Logger(pkg.Trace("main.seedFile", "main"))
  logger.Info().
    Str("name", name).
    Msg("Enter")

  // get md5 sum of name, to avoid collisions with
  // existing tmpdir files
  hash := fmt.Sprintf("%x", md5.Sum([]byte(name)))
  logger.Info().
    Str("name", name).
    Str("hash", hash).
    Msg("Determined md5 hash of name")

  // create a seed file required to create an ipc
  // int
  path := fmt.Sprintf("%s/mq-%s", os.TempDir(), hash)
  return path
}
