#!/bin/sh

systemctl stop arclift.service \
&& systemctl disable arclift.service \
&& rm -f /etc/systemd/system/arclift.service \
&& systemctl daemon-reload
