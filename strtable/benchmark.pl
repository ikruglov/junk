#!/usr/bin/env perl

use strict;
use warnings;

use Bench;
use Find::Lib 'lib', 'blib', 'blib/arch/';
use Benchmark qw(cmpthese :hireswallclock);

open(my $fh, '<', $ARGV[0]) or die $!;
my @corpus = <$fh>;
close($fh);

cmpthese(-10, {
    stbl_lvlhash_1K      => sub { Bench::benchmark_strtable_leveldb_hash(\@corpus, 1024) },
    stbl_lvlhash_64K     => sub { Bench::benchmark_strtable_leveldb_hash(\@corpus, 64 * 1024) },
    stbl_lvlhash_256K    => sub { Bench::benchmark_strtable_leveldb_hash(\@corpus, 256* 1024) },
    stbl_lvlhash_1M      => sub { Bench::benchmark_strtable_leveldb_hash(\@corpus, 1024 * 1024) },
    stbl_perlhash_1K     => sub { Bench::benchmark_strtable_perl_hash(\@corpus, 1024) },
    stbl_perlhash_64K    => sub { Bench::benchmark_strtable_perl_hash(\@corpus, 64 * 1024) },
    stbl_perlhash_256K   => sub { Bench::benchmark_strtable_perl_hash(\@corpus, 256 * 1024) },
    stbl_perlhash_1M     => sub { Bench::benchmark_strtable_perl_hash(\@corpus, 1024 * 1024) },
    native_hv            => sub { Bench::benchmark_hv(\@corpus) },
});
