#!/bin/sh
if [ ! -f stubby4j.jar ]; then
    wget -c "https://search.maven.org/remotecontent?filepath=io/github/azagniotov/stubby4j/6.0.1/stubby4j-6.0.1.jar" -O stubby4j.jar
fi
java -jar stubby4j.jar -d github-api.yml -l 0.0.0.0
