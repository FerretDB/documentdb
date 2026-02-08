/*-------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation.  All rights reserved.
 *
 * include/commands/coll_mod.h
 *
 * Exports around the functionality of collmod
 *
 *-------------------------------------------------------------------------
 */
#ifndef DOCUMENTDB_COLL_MOD_H
#define DOCUMENTDB_COLL_MOD_H

void UpdatePostgresIndexCore(uint64_t collectionId, int indexId, bool hidden, bool
							 ignoreMissingShards);

#endif
