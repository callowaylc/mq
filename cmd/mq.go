package main

// imports //////////////////////////////////////

import (
  "os"
  "fmt"
  "C"

  "github.com/spf13/cobra"
  "github.com/siadat/ipc"


  "github.com/callowaylc/mq/pkg"
  "github.com/callowaylc/mq/pkg/log"
)

// constants ////////////////////////////////////

const ExitStatusArgument int = 3
const ExitStatusSysv int = 40
const ExitStatusSysvRead int = 41
const ExitStatusSysvWrite int = 42


// main /////////////////////////////////////////

func init() {
  log.Init()
}

func main() {
  logger := log.Logger(pkg.Trace("main.main", "main"))
  logger.Info().Msg("Enter")
  defer logger.Info().Msg("Exit")

  var fsizef bool
  var fdelete bool

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
        key, err := ipc.Ftok(args[0], 42)
        if err != nil {
          logger.Error().
            Str("path", args[0]).
            Str("error", err.Error()).
            Msg("Failed to generate key")
          os.Exit(ExitStatusArgument)
        }
        logger.Info().
          Int64("key", int64(key)).
          Str("path", args[0]).
          Msg("Created IPC key")

        // attempt to create queue
        qid, err := ipc.Msgget(key, ipc.IPC_CREAT)
        if err != nil {
          logger.Error().
            Str("path", args[0]).
            Int64("qid", int64(qid)).
            Str("error", err.Error()).
            Msg("Failed to create queue")
          os.Exit(ExitStatusArgument)
        }
        logger.Info().
          Int64("qid", int64(qid)).
          Msg("Obtained queue identifier")

        if fdelete {
          logger.Info().
            Int64("qid", int64(qid)).
            Str("path", args[0]).
            Msg("Signal queue deletion")

          // if delete flag has been specified, defer deltion
          defer func() {
            err := ipc.Msgctl(qid, ipc.IPC_RMID)
            if err != nil {
              logger.Error().
                Int64("qid", int64(qid)).
                Str("path", args[0]).
                Str("error", err.Error()).
                Msg("Failed to delete queue")
              os.Exit(ExitStatusSysv)
            }
            logger.Info().
              Int64("qid", int64(qid)).
              Str("path", args[0]).
              Msg("Queue deleted")
          }()

        } else if fsize {
          mqds := C.struct_msqid_ds{}
          rc, err := C.msgctl(C.int(qid), C.IPC_STAT, &mqds)

          if rc == -1 {
            logger.Error().
              Int64("qid", int64(qid)).
              Str("path", args[0]).
              Str("error", "Failed something").
              Msg("Failed to stat queue")
            os.Exit(ExitStatusSysv)
          }

          fmt.Println(mqds.msg_qnum)


        } else {
          // otherwise, determine if read or write operation
          // against queue
          if len(args) == 1 {
            // we are performing a read operation
            logger.Info().
              Int64("qid", int64(qid)).
              Str("path", args[0]).
              Msg("Performing read operation")

              msg := &ipc.Msgbuf{Mtype: 12}
              err := ipc.Msgrcv(qid, msg, 0)
              if err != nil {
                logger.Error().
                  Str("error", err.Error()).
                  Msg("Failed to read message")
                os.Exit(ExitStatusSysvRead)
              }

              // write message to stdout
              fmt.Println(string(msg.Mtext))

          } else if len(args) == 2 {
            // we are performing a write operation
            logger.Info().
              Int64("qid", int64(qid)).
              Str("path", args[0]).
              Str("payload", args[1]).
              Msg("Performing write operation")


            msg := &ipc.Msgbuf{Mtype: 12, Mtext: []byte(args[1])}
            err := ipc.Msgsnd(qid, msg, 0)
            if err != nil {
              logger.Error().
                Str("error", err.Error()).
                Msg("Failed to send message")
              os.Exit(ExitStatusSysvWrite)
            }

          } else {
            // log wasted arguments - this can be useful when
            // interacting with the binary from a shell environment
            err := ipc.Msgctl(qid, ipc.IPC_STAT)
            if err != nil {
              logger.Error().
                Str("error", err.Error()).
                Msg("Uhhok")
              os.Exit(ExitStatusSysv)
            }

            logger.Error().
              Str("args", fmt.Sprint(args)).
              Msg("Too many messages")
            os.Exit(ExitStatusArgument)
          }
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
    &fid, "size", "s", false, "Queue size",
  )
  root.PersistentFlags().BoolVarP(
    &fdelete, "delete", "d", false, "Delete queue",
  )
  root.Execute()
}
