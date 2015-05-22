#!/usr/bin/env perl

use strict;
use warnings;

use Bench;
use Find::Lib 'lib', 'blib', 'blib/arch/';
use Benchmark qw(cmpthese :hireswallclock);

open(my $fh, '<', $ARGV[0]) or die $!;
my @corpus = <$fh>;
close($fh);

cmpthese(10, {
    strtable_8  => sub { Bench::benchmark_strtable(\@corpus, 8) },
    strtable_1K => sub { Bench::benchmark_strtable(\@corpus, 1024) },
    strtable_1M => sub { Bench::benchmark_strtable(\@corpus, 1024 * 1024) },
    native => sub { Bench::benchmark_hv(\@corpus) },
});
