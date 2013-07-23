onetime
=======

One-Time file sharer

# What is this?

*onetime* is an HTTP-based tool meant for file sharing from an Internet
box. Select a file for sharing and get a temporary URL that can be clicked
only once for download. Share the URL with the intended recipient. Done.

The problem I wanted to solve is simple: I rent a Debian box from a hosting
provider with 1 TB of disk storage. This box is mainly used to store family
pictures and such, and I wanted to be able to share large files with
friends and relatives without having to give them full access to the box or
starting an insecure ftp server or equivalent. Since most recipients are
barely computer-literate, the sharing solution has to be really simple.
Sending them a link to click seemed like an obvious idea.

I tried several solutions before that and found none to be really
satisfactory. A properly-configured *lighttpd* could do the job but I could
not get past the Lua scripting part, though I happen to know Lua pretty
well. My next try was a Python-based app (with web.py) but it still relied
on *lighttpd* for file service, and the Python process took enormous
amounts of CPU and RAM while serving files.

Anyway, I am still learning Go and love the language. Implementing this in
Go was a breeze, is contained in a single file, and solves both the
command-line part and the HTTP(s) server.

Fork at will! Pull requests welcome!

# How to use it?

*onetime* is purely command-line driven. Launch it without arguments to get
some help. Commands are:

    config          configure server
    server          start the server in background
    add FILENAME    create a new one-time token for FILENAME
    ls              list existing entries
    del TOKEN       delete a one-time token

The server part can be started/stopped on Debian using standard init.d
scripts. One is provided here as an example.


