#!/bin/sh

arclift config init \
&& systemctl daemon-reload \
&& systemctl enable arclift.service \
&& systemctl start arclift.service
