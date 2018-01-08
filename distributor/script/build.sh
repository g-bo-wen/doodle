#!/bin/bash

set -e

url=$1
key=$2

framework_package="github.com/dearcode/doodle/service"

function convert_url() {
    if [[ "$url" =~ "http://" ]]
    then
        url=`echo $url|sed 's/http:\/\//git@/g'|sed 's/\//:/'|sed 's/$/.git/'`
    fi
}

function create_path() {
    base_path=`echo $url|awk -F'[@:/]' '{print "src/"$2"/"$3}'`
    rm -rf $base_path
    mkdir -p $base_path
}

function clone_source() {
    cd $base_path;
    git clone --depth=1 $url;
    cd -;

    app=`echo $url|xargs basename -s .git`;
    base_path=$base_path/$app;

    cd $base_path;

    git_hash=`git log --pretty=format:'%H' -1`
    git_time=`git log --pretty=format:'%ci' -1`
    git_message=`git log --pretty=format:'%cn %s %b' -1`

    rm -rf .git

    cd -;
}

function create_dockerfile() {
    project=`echo $url|sed 's/.*@//'|sed 's/\.git//'|sed 's/:/\//'`;
    package_in_vendor="$project/vendor/$framework_package"
    cp Dockerfile.tpl Dockerfile
    sed -i "s#{{BASE_PATH}}#$base_path#" Dockerfile
    sed -i "s#{{PACKAGE_IN_VENDOR}}#$package_in_vendor#g" Dockerfile
    sed -i "s#{{KEY}}#$key#g" Dockerfile
    sed -i "s#{{GIT_HASH}}#$git_hash#g" Dockerfile
    sed -i "s#{{GIT_TIME}}#$git_time#g" Dockerfile
    sed -i "s#{{GIT_MESSAGE}}#$git_message#g" Dockerfile
    sed -i "s#{{PROJECT}}#$project#g" Dockerfile
}

function build() {
    version=`date -d "$git_time" +%Y%m%d.%H%M`
    image="$project:$version"
    docker build --no-cache -t $image .
    docker run -i --rm -v $PWD/bin:/base $image bash -c 'cp $GOPATH/bin/* /base/' 
}


convert_url

create_path

clone_source

create_dockerfile

build
