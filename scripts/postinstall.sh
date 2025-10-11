#!/bin/sh

gos3backup config init \
&& systemctl daemon-reload \
&& systemctl enable gos3backup.service \
&& systemctl start gos3backup.service
