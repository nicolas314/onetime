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

You can read a bit more from this blog post:
http://nicolas314.wordpress.com/2013/07/24/one-time-file-sharing

Fork at will! Pull requests welcome!

# How to build

    go build onetime.go


# How to use

*onetime* is purely command-line driven. Launch it without arguments to get
some help. Commands are:

    onetime config          Configure server
    onetime serve           Serve onetime requests
    onetime add path        Create onetime request for path
    onetime ls              List existing requests
    onetime del token       Delete onetime request
    onetime purge           Delete all expired tokens


- config will create a default configuration file called onetime.json in
  the same directory as the onetime executable. Edit this file before
  launching anything else

- server starts the program in server mode. The server remains in the
  foreground while running. You can transform that into a background daemon
  on Debian e.g. by using start-stop-daemon.

- add registers a file for service. It prints out on stdout a short
  message meant to be copied/pasted into an email. The file name can be
  provided with full path. Without path indication, onetime will search the
  current working directory for a matching file name.

- ls lists all onetime tokens currently registered

- del token removes a token from the DB. A token in that case is the 8-char
  random string generated for each file.

- purge removes all tokens that have expired, i.e. have been clicked
  more than 4 hours ago.


The server part can be started/stopped on Debian using standard init.d
scripts. One is provided here as an example: see onetimed.

Files are served directly by the Go process, using the default HTTP server
implementation from Go. Files are served on HTTP by default. To switch to
HTTPS, indicate a certificate and key file name in the json configuration
file. Example:

    {
        "TOKEN_DB": "token.db",
        "LOG_FILE": "onetime.log",
       "BASE_ADDR": "http|https://FQDN:PORT",
             "CRT": "server.crt",
             "KEY": "server.key"
    }

CRT and KEY are not necessary for HTTP service, only HTTPS.

The json configuration file is called onetime.json and must live in the
same directory as the executable file.

Other configuration file names provided without path (e.g. token.db) are
expected in the same directory as the executable file. If you want to put
them somewhere else, indicate a full path to access them, e.g.
/var/onetime/token.db.

BASE_ADDR is actually a URL. It should point to an address that is visible
from your intended audience. Examples:

 - http://myhost.example.com:1234
 - https://myhost.example.com:2500

Careful about indicating http or https in the URL. If you want to serve
over HTTPS you need to have a certificate and key for the server.

CRT and KEY are X.509 certificate and key files. Do not protect the key
file with a password if you want the server to start without interaction.


# More details

There are few Linuxisms in the code: paths are all slash-separated,
and the configuration file is found by searching for *onetime.json* in the
same directory as the executable, found by parsing /proc/self/exe.

# Wish list

A few things would be worth concentrating on:

- Used tokens are not deleted automatically. They should.
- Adding the possibility to share an entire directory: the sharing page
  should then show one link per file and an additional link to download all
  above files as a single zip.
- The sharing page could be i18n'd.
- The CSS could use better design
- An admin page could be added to monitor current tokens from a web UI,
  review logs, etc.
- One-time tokens could be mailed directly from the program. Right now they
  are just printed out on stdout.


