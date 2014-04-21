#!/usr/bin/perl

use strict;
use warnings;
use Data::Dumper;
use Sereal::Encoder;
use Sereal::Decoder;
use Benchmark qw(:all);

open(my $fh, '<', $ARGV[0]) or die $!;
$/ = undef;
my $content = <$fh>;
close $fh;

my $array = Sereal::Decoder->new()->decode($content);
print "cnt: " . @$array . "\n";

my $enc = Sereal::Encoder->new()->encode($array);
print "enc: " . bytes::length($enc) . "\n";

my $enc_snappy = Sereal::Encoder->new({ snappy => 1 })->encode($array);
printf "enc_snappy: %d (%.2f)\n", bytes::length($enc_snappy), bytes::length($enc_snappy) / bytes::length($enc);

my $enc_snappy_inc = Sereal::Encoder->new({ snappy_incr => 1 })->encode($array);
printf "enc_snappy_inc: %d (%.2f)\n", bytes::length($enc_snappy_inc), bytes::length($enc_snappy_inc) / bytes::length($enc);

my $enc_snappy_dedup = Sereal::Encoder->new({ snappy_incr => 1, dedupe_strings => 1 })->encode($array);
printf "enc_snappy_dedup: %d (%.2f)\n", bytes::length($enc_snappy_dedup), bytes::length($enc_snappy_dedup) / bytes::length($enc);

cmpthese(100, {
    'enc'              => sub { Sereal::Encoder->new()->encode($array) },
    'enc_snappy'       => sub { Sereal::Encoder->new({ snappy_incr => 1 })->encode($array) },
    'enc_snappy_inc'   => sub { Sereal::Encoder->new({ snappy_incr => 1 })->encode($array) },
    'enc_snappy_dedup' => sub { Sereal::Encoder->new({ snappy_incr => 1, dedupe_strings => 1 })->encode($array) },
});
exit 0;
