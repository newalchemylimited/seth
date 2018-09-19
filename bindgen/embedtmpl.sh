#!/bin/sh

cat <<EOT > template.go
package main

// Don't edit this file. 

var tmpl = \`$(cat tmpl.gotmpl | sed -e 's/`/`+"`"+`/')\`
EOT