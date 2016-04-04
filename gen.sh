#!/bin/bash

git show master:README.md|marked --gfm --breaks > readme.html
perl -pi -e 's/<li>\[ \] /<li class="no-bullet"><input type="checkbox" class="task-list-item-checkbox" checked="checked" disabled="disabled">\&nbsp;/g,
             s/<li>\[x\] /<li class="no-bullet"><input type="checkbox" class="task-list-item-checkbox" disabled="disabled">\&nbsp;/g' readme.html
sed -e '/<!-- Content -->/r readme.html' readme.tmpl > index.html
rm -f readme.html
