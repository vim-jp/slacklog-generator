#!/bin/bash
if [ $docker -ne 0 ] ; then
  docker run --rm -t --volume="$PWD:/srv/jekyll" jekyll/jekyll:pages jekyll build -d $1 > $1.docker-jekyll-build.log 2>&1
else
  jekyll build -d $1 > $1.jekyll-build.log 2>&1
fi
