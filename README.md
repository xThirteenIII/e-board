# Simple Shell in Go
## A POSIX-like base shell written in Go, to be used as a support for an electronic chessboard in the future.

**CODE IT'S ALL WRITTEN BY ME**

### **Why a shell?**
I've always been in love with the terminal and its job.
Recently I read "Just for Fun" by Linus Torvald, and I was fascinated by how he spent 5 months just to build Linux 0.01. This included a terminal emulator and all the kernel. So to learn more about OS and how things work at low level, I decided to build a basic shell in Go to better understand.

### **Why a chessboard?**
I'm a huge fan of Harry potter and watching the first one again for the 1Mth time, I had the idea to make an  
automated self-moving board. Unfortunately that turns out to be expensive, so I'll just start with a normal chessboard  
with sensors to process moves and send them to an online server (e.g. Chess.com).

### **How am I building it?**
I've started to read Computer Systems, A Programmer's Prospective. There's all I need. I'm **NOT** using Claude or GPT
because for me that's the only way not to learn stuff. Ofc i'm not a fool so **I AM** using GPT for knowing where to
look and to make the process faster, but i still like StackOverflow better. 

### **Where i'm at so far**
I have implemented basic built-in commands such as 'echo', 'pwd', 'cd', 'exit', 'type'. 
For external commands, I use the exec Go package. I tried to use syscall package too, but syscall.ForkExec() felt like too much headaches to deal with at the moment. The exec.Command{} and exec.Run() / Start() functions give me all flexibility I need with a nice wrapper. Each command line is parsed into a slice of command + operatorAfter, that is the operator, if any, that follows a command (&, |, >, <). At the moment, I'm arguing with processes, PIDs and syscall.SIGNALS to handle background jobs and being sure not to have zombie processes around.
