#!/bin/sh
# Lowest CPU priority, and idle disk priority when the kernel allows it.
# Tesseract and pdftoppm inherit both, so OCR never slows DEEIX down.
if ionice -c 3 true 2>/dev/null; then
  exec ionice -c 3 nice -n 19 /usr/local/bin/deeix-ocr "$@"
fi
exec nice -n 19 /usr/local/bin/deeix-ocr "$@"
