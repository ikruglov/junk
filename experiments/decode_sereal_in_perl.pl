#!/usr/bin/perl

use strict;
use warnings;
use Data::Dumper;
use Sereal::Decoder;

open(my $fh, '<', $ARGV[0]) or die $!;
$/ = undef;
my $content = <$fh>;
close $fh;

my $srl = Sereal::Decoder->new();
my $array = $srl->decode($content);
print @$array . "\n";
exit 0;
