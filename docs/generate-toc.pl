#!/usr/bin/perl

while($line=<STDIN>) {
    chomp $line;

    if($line =~ /^#/) {
	# We have a section heading
	#$indent = $line =~ tr/#//;
	$indent = $line;
	$indent =~ s/[^#]//g;	# remove all non-hash chars
	$indent =~ s/#//;	# Eat the first one (ie, no indentation)
	$indent =~ s/#/   /g;	# Indent based on the number of hashes
	$line =~ s/#//g;
	$anchor = lc($line);
	$anchor =~ s/[^a-zA-Z0-9]//g;
	print "${indent}1. [$line](#$anchor)\n";
    }

    #print "$line\n";
}

