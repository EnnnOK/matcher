c	matches any literal character c
.	matches any single character
*	matches zero or more occurrences of the previous character
|	matches the previous character or the next character

This regular expression package is inspired from Russ Cox's
article "Regular Expression Matching Can Be Simple And Fast".
https://swtch.com/~rsc/regexp/regexp1.html

Uses Thompson construction for the NFA from
Communications of the ACM
Volume 11 Issue 6, June 1968
"Programming Techniques: Regular expression search algorithm"

DFA construction is based off of powerset construction while
caching the result after each state.

Lexer is inspired by Rob Pike's lecture: "Lexical Scanning in Go"
https://youtu.be/HxaD_trXwRE