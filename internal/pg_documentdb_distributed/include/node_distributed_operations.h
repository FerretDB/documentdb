/*-------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation.  All rights reserved.
 *
 * include/node_distributed_oeprations.h
 *
 * The implementation for node level distributed operations.
 *
 *-------------------------------------------------------------------------
 */
#ifndef DOCUMENTDB_NODE_DISTRIBUTED_OPS_H
#define DOCUMENTDB_NODE_DISTRIBUTED_OPS_H

void UpdateDistributedPostgresIndex(uint64_t collectionId, int indexId, bool hidden);

#endif
