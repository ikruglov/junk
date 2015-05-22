#ifndef _STRTABLE_H_
#define _STRTABLE_H_

#include <stdlib.h>
#include <string.h>

// Implementation of leveldb_hash (with minor changes) if taken from 
// https://github.com/google/leveldb/blob/master/util/hash.cc
//
// Copyright (c) 2011 The LevelDB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file. See the AUTHORS file for names of contributors.

// The FALLTHROUGH_INTENDED macro can be used to annotate implicit fall-through
// between switch labels. The real definition should be provided externally.
// This one is a fallback version for unsupported compilers.
#ifndef FALLTHROUGH_INTENDED
#define FALLTHROUGH_INTENDED do { } while (0)
#endif

inline uint32_t leveldb_hash(const char* data, size_t n) {
    // Similar to murmur hash
    const uint32_t seed = 0xbc9f1d34;
    const uint32_t m = 0xc6a4a793;
    const uint32_t r = 24;
    const char* limit = data + n;
    uint32_t h = seed ^ (n * m);

    // Pick up four bytes at a time
    while (data + 4 <= limit) {
        uint32_t w = (uint32_t) data;
        data += 4;
        h += w;
        h *= m;
        h ^= (h >> 16);
    }

    // Pick up remaining bytes
    switch (limit - data) {
        case 3:
            h += ((unsigned char) data[2]) << 16;
            FALLTHROUGH_INTENDED;
        case 2:
            h += ((unsigned char) data[1]) << 8;
            FALLTHROUGH_INTENDED;
        case 1:
            h += (unsigned char) data[0];
            h *= m;
            h ^= (h >> r);
            break;
    }
    return h;
}

struct strtable_element {
    uint32_t hash;
    uint32_t value;
    uint32_t keylen;
    const char *key;
};

struct strtable {
    uint32_t keys;
    uint32_t capacity;          // has to be power of 2
    struct strtable_element *t;
};

struct strtable_result {
    uint32_t value;
    uint32_t inserted;
};

typedef struct strtable strtable_t;
typedef struct strtable_result strtable_result_t;
typedef struct strtable_element strtable_element_t;

/* NOTE!!!!!
 * Library crashes app upon malloc failure
 * */

inline void strtable_init(strtable_t *tbl, uint32_t capacity)
{
    tbl->keys = 0;
    tbl->capacity = capacity;
    tbl->t = (strtable_element_t*) calloc(capacity, sizeof(strtable_element_t));
    if (!tbl->t) abort();
}

inline void strtable_clear(strtable_t *tbl)
{
    free(tbl->t);
}

inline void strtable_grow(strtable_t *tbl)
{
    uint32_t newslot;
    uint32_t newcap = tbl->capacity * 2;
    uint32_t newmask = newcap - 1;

    strtable_element_t *t    = tbl->t;
    strtable_element_t *newt = (strtable_element_t*) calloc(newcap, sizeof(strtable_element_t));
    if (!newt) abort();

    for (uint32_t i = 0; i < tbl->capacity; ++i) {
        if (t[i].key == NULL) continue;

        newslot = t[i].hash & newmask;

        while (1) {
            if (newt[newslot].key == NULL) {
                newt[newslot] = t[i];
                break;
            }

            newslot = (newslot + 1) & newmask;
        }
    }

    tbl->capacity = newcap;
    tbl->t = newt;
}

inline strtable_result_t strtable_insert(strtable_t *tbl, const char *str, uint32_t len, uint32_t value)
{
    uint32_t mask = tbl->capacity - 1;
    uint32_t hash = leveldb_hash(str, len);
    int slot = hash & mask;

    strtable_result_t res = { value, 0 };
    strtable_element_t *t = tbl->t;

    while (1) {
        if (t[slot].key == NULL) {
            t[slot].hash = hash;
            t[slot].value = value;
            t[slot].key = str;
            t[slot].keylen = len;

            tbl->keys++;
            if (tbl->keys > (tbl->capacity >> 1)) // tbl->capacity / 2
                strtable_grow(tbl);

            res.inserted = 1;
            return res;
        }

        if (   t[slot].hash == hash
            && t[slot].keylen == len
            && strncmp(t[slot].key, str, len) == 0
        ) {
            return res;
        }

        slot = (slot + 1) & mask;
    }
}

#endif
