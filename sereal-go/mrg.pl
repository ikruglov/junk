#!/usr/bin/env perl

use strict;
use warnings;

use Time::HiRes;
use Sereal::Encoder;
use Sereal::Decoder;

my @data;
my @files = glob("github.com/Sereal/Sereal/Go/sereal/data/*.srl");
foreach my $file (@files) {
    $/ = undef;
    open(my $fh, '<', $file) or die $!;
    my $val = <$fh>;
    push @data, $val;
}

my $start = Time::HiRes::time();

my $encoder = Sereal::Encoder->new({ dedupe_strings => 1, compress => 1 });
my $decoder = Sereal::Decoder->new();

my @events = map { $decoder->decode($_) } @data;
print $encoder->encode(\@events);

printf STDERR "The call took %.2fs to run\n", Time::HiRes::time - $start;
exit(0);

