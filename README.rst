cellaserv2
==========

RPC broker based on protobuf messages written in Go.

Features:

- Request-Reply
- Publish-Subscribe
- Log messages to pcap

Should be used in conjunction with `cellaservctl
<https://bitbucket.org/evolutek/cellaservctl>`_.

Install
-------

After having installed go and set GOPATH, run:

    $ go get bitbucket.org/evolutek/cellaserv2

Start
-----

Run, with $GOPATH/bin in your PATH:

    $ cellaserv2

Client libraries
----------------

- `python-cellaserv2 <https://bitbucket.org/evolutek/python-cellaserv2>`_
  Python3 library

Authors
-------

- Rémi Audebert, evolutek<< 2014
