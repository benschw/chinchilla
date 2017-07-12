FROM ubuntu:16.04

ADD ./chinchilla /chinchilla

ENTRYPOINT /chinchilla
