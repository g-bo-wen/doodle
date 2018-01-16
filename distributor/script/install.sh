#!/bin/bash


project=$1
pid=$1
port=$2
etcd=$3


function project_init() {
    mkdir -p $project/bin
    mkdir -p $project/logs
    mkdir -p $project/config
}


function send_close_signal() {
    pgrep_pid=`pgrep $project`
    port_pid=`netstat -ntpl|awk "/$port/{print $NF}"|awk -F/ '{print $1}'`

    if [ $pgrep_pid == $port_pid ];
    then
        kill -s SIGUSR1 $pid
    fi
}


function extract_project() {
    tar xzf -C bin $project.tar.gz; 
}

function daemon_start() {
    nohup ./bin/$project -etcd $etcd -h : >> daemon.log 2>&1 &
}


cd $project

project_init

extract_project

send_close_signal

daemon_start

cd -
