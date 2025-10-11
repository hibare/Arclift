#!/bin/sh

systemctl stop gos3backup.service \
&& systemctl disable gos3backup.service \
&& rm -f /etc/systemd/system/gos3backup.service \
&& systemctl daemon-reload
