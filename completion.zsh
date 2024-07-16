#compdef goes

if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
	COMPREPLY=( $(goes complete ${COMP_WORDS[@]:1} '') )
else
	COMPREPLY=( $(goes complete ${COMP_WORDS[@]:1}) )
fi

return 0
