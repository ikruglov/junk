#!/bin/env perl

use strict;
use warnings;

BEGIN {
    push @INC, "/Users/ikruglov/src/perl/Sereal/Perl/Encoder/blib/lib/";
    push @INC, "/Users/ikruglov/src/perl/Sereal/Perl/Encoder/blib/arch";
    push @INC, "/Users/ikruglov/src/perl/Sereal/Perl/Decoder/blib/lib/";
    push @INC, "/Users/ikruglov/src/perl/Sereal/Perl/Decoder/blib/arch";
}

use Time::HiRes;
use Sereal::Encoder;
use Sereal::Decoder;
use Getopt::Long qw(GetOptions);

my $file = '';
my $repeat = 1;
my $dump_to;
my $dedupe_strings = 0;

GetOptions(
    'file=s'                        => \$file,
    'repeat=i'                      => \$repeat,
    'dump_to|dump-to=s'             => \$dump_to,
    'dedupe_strings|dedupe-strings' => \$dedupe_strings,
);

print "option:\n";
print "repeat: $repeat\n";
print "dedupe-strings: " . ($dedupe_strings // '') . "\n";
print "dump-to: " . ($dump_to // '') . "\n";
print "file: $file\n";
print "\n";

my $decoder = Sereal::Decoder->new();
my $encoder = Sereal::Encoder->new({
    dedupe_strings => $dedupe_strings
});

$/ = undef;
my @data;

my @files = -d $file ? glob("$file/*.srl") : $file;
print "read and deserialize " . scalar @files . " files\n";
scalar @files == 0 and die "no files to read\n";
foreach my $file (@files) {
    open(my $fh, '<', $file) or die $!;
    my $content = <$fh>;
    close($fh);

    my $decoded = $decoder->decode($content);
    my $ref = ref $decoded;

    if ($ref eq 'ARRAY') {
        push @data, @$decoded;
    } else {
        push @data, $decoded;
    }
}

my $result;
my @timings;
print "serialize data (array: " . scalar @data . " repeat: $repeat)\n";

foreach (1..$repeat) {
    my ($swall, $scpu) = __times();
    $result = $encoder->encode(\@data);
    my ($ewall, $ecpu) = __times();

    push @timings, {
        wall => $ewall - $swall, 
        cpu  => $ecpu  - $scpu,
    };

    printf("call took wall: %.2f sec; cpu: %.2f sec\n", $timings[-1]->{wall}, $timings[-1]->{cpu});
}

foreach my $t ('wall', 'cpu') {
    my %stats = __stats(map { $_->{$t} } @timings);
    printf("stats %-4s avg %.2f sec; stddev %.2f sec; min %.2f; med %.2f; max %.2f\n",
           $t, $stats{avg}, $stats{stddev}, $stats{min}, $stats{med}, $stats{max});
}

if ($dump_to) {
    print "Dumping result to $dump_to\n";
    open(my $fh, '>', $dump_to) or die $!;
    print $fh $result;
    close $fh;
}

exit 0;

sub __times {
    my $w = Time::HiRes::time;
    my ($u, $s) = times();
    return ($w, $u + $s);
}

sub __stats {
    # The caller is supposed to have done this sorting
    # already, but let's be wasteful and paranoid.
    my @v = sort { $a <=> $b } @_;
    my $min = $v[0];
    my $max = $v[-1];
    my $med = @v % 2 ? $v[@v/2] : ($v[@v/2-1] + $v[@v/2]) / 2;
    my $sum = 0;
    for my $t (@_) {
        $sum += $t;
    }
    my $avg = $sum / @_;
    my $sqsum = 0;
    for my $t (@_) {
        $sqsum += ($avg - $t) ** 2;
    }
    my $stddev = sqrt($sqsum / @_);
    return ( avg => $avg,
             stddev => $stddev,
             rstddev => $avg ? $stddev / $avg : undef,
             min => $min, med => $med, max => $max );
}
