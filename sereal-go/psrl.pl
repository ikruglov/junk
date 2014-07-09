#!/usr/local/bin/booking-perl

use strict;
use warnings;

use Time::HiRes;
use Sereal::Encoder;
use Sereal::Decoder;

my $decoder = Sereal::Decoder->new();
my $encoder = Sereal::Encoder->new();

$/ = undef;
my @data;

print "read and deserialize data\n";
foreach my $file (glob($ARGV[0] . "/*.srl")) {
    open(my $fh, '<', $file) or die $!;
    my $content = <$fh>;
    close($fh);

    push @data, $decoder->decode($content)
}

print "serialize data\n";
my $start = Time::HiRes::time();
my $result = $encoder->encode(\@data);
printf("Call tool %.2f seconds\n", Time::HiRes::time() - $start);
