#!/usr/bin/perl

use strict;
use warnings;
use Data::Dumper;
use Sereal::Encoder;
use Sereal::Decoder;

open(my $fh, '<', $ARGV[0]) or die $!;
$/ = undef;
my $content = <$fh>;
close $fh;

my $array = Sereal::Decoder->new()->decode($content);
print @$array . "\n";

my $enc = Sereal::Encoder->new()->encode($array);
print "enc: " . bytes::length($enc) . "\n";

my $enc_snappy = Sereal::Encoder->new({ snappy => 1 })->encode($array);
print "enc_snappy: " . bytes::length($enc_snappy) . "\n";

my $enc_snappy_inc = Sereal::Encoder->new({ snappy_incr => 1 })->encode($array);
print "enc_snappy_inc: " . bytes::length($enc_snappy_inc) . "\n";

my $enc_snappy_dedup = Sereal::Encoder->new({ snappy_incr => 1, dedupe_strings => 1 })->encode($array);
print "enc_snappy_dedup " . bytes::length($enc_snappy_dedup) . "\n";
exit 0;
