#!/bin/bash

git show master:README.md|marked --gfm > readme.html
sed -e '/<!-- Content -->/r readme.html' readme.tmpl > index.html
rm -f readme.html
