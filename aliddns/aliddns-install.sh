#!/bin/bash

installServer(){
  if [ ! -d /home/ddns ];then
    useradd ddns
  fi
  mkdir -p /home/ddns/aliddns
  cp -rf * /home/ddns/aliddns/
  chown -R ddns:ddns /home/ddns
  if [ ! -f /etc/systemd/system/aliddns.service ];then
    cp -f aliddns.service /etc/systemd/system/
  fi
  mkdir -p /home/ddns/conf/
  if [ ! -f /home/ddns/conf/aliddns.properties ];then
    cp -f conf/aliddns.properties /home/ddns/conf/aliddns.properties
  fi
  systemctl enable aliddns.service
}

case "$1" in
  -i)
    installServer
    ;;
  *)
    echo "Usage: ./aliddns-install.sh -i"
    ;;
esac