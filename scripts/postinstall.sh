#!/bin/sh

mkdir -p /etc/arclift \
&& touch /etc/arclift/config.yaml \
&& arclift config init \
&& systemctl daemon-reload \
&& systemctl enable arclift.service \
&& systemctl start arclift.service
