# mq

A simple command-line message queue, intended as an [IPC](https://www.tldp.org/LDP/tlk/ipc/ipc.html) utility between shell processes, built on top of [sysv message queues](https://www.softprayog.in/programming/interprocess-communication-using-system-v-message-queues-in-linux) and [shopify's sysv wrapper](https://github.com/Shopify/sysv_mq).

* [Introduction](#introduction)
* [Installation](#installation)
* [Usage](#usage)

## Introduction

Operational work, that is heavily invested in shell scripting, will inevitably run across the need to provide queueing logic, as a means of coordination between parallel processes, whether they be disparate or forks of a common parent. Coordinating "thread" access, to shared resources, using posix supplied mechanisms, is awkward at best:

  - [FIFO Pipes](https://www.gnu.org/software/libc/manual/html_node/Pipes-and-FIFOs.html) are best used as a one-in, one-out structures, that require many hoops to support multiple consumer/producer models, that adhere to some form of message order and still provide selective blocking.
  - Files require mutex locks, using tools like [flock](https://linux.die.net/man/1/flock) or custom logic, are "user-space" bound (as opposed purely kernel managed) and awareness has to be dedicated to state and clean.
  - Local Sockets work as IPC mechanism, but are awkward to fit into a queue paradigm.
  - External queues, like [redis](https://redis.io), obviously require expanding a given stack and thus increased operational complexity.

The `mq` tool acts as an interface to sysv message queues, which are kernel managed and available on most \*nix platforms, including BSD; message order is guarenteed, reads will block on (lack of)availability and \*[perfomance will exceed]() "user space" queueing systems, for those of you that care.

\**`mq` is not intended as a performant bus framework*

## Installation

Binary [releases](./releases) are available for Darwin/Linux AMD64 targets, with no plans to provide additional platform support; feel free review the [Makefile](./Makefile) if you wish to compile for a given platform.

## Usage

```bash
$ ./build/mq -h
Usage:
  mq key [value] [flags]

Flags:
  -c, --count    Number of messages in the queue
  -d, --delete   Delete the queue
  -h, --help     help for mq
```

## Examples

\* Write to queue `asdf`.

```bash
$ ./build/mq asdf one
$ echo $?
#$ ipcs -qo
IPC status from <running system> as of Thu Apr 25 14:29:50 EDT 2019
T     ID     KEY        MODE       OWNER    GROUP CBYTES  QNUM
Message Queues:
q 393216 0x2a042aa4 --rw-rw---- christiancalloway    staff      3      1
```

\* Write to queue `asdf` from stdin and read.

```bash
$ cat <<eof | ./build/mq asdf
> one
> two
eof
$ ./build/mq asdf
one
$ ./build/mq asdf
two
$ echo $?
0
```

\* Block until message is available.

```bash
$ ( sleep 5 && ./build/mq asdf delay ) & gtime -f '%e' ./build/mq asdf
[1] 42090
delay
5.01
[1]+  Done                    ( sleep 5 && ./build/mq asdf delay )
```

\* Get the message count.

```bash
$ ./build/mq asdf astring
$ ./build/mq asdf --count
1
$ ./build/mq asdf
astring
$ ./build/mq asdf --count
0
```

The last two examples should make clear that its up to the implemntor to manage read blocks.
