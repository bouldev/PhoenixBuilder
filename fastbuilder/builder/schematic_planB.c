// +build do_not_add_this_tag_

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <zlib.h>
#ifndef WIN32
#include <arpa/inet.h>
#else
#include <winsock.h>
#endif
#include <errno.h>
#include <string.h>

#define ERR_INVALID_SCHEMATIC_FILE 1
#define ERR_NOT_IMPLEMENTED 2
#define ERR_EARLY_TAG_END 3
#define ERR_UNEXPECTED_TAG_TYPE 4
#define ERR_NOT_FOUND 5

static void goIgnoreTagName(gzFile file) {
  uint16_t strLen;
  gzread(file, &strLen, 2);
  strLen = htons(strLen);
  // printf("%d\n",strLen);
  gzseek(file, strLen, SEEK_CUR);
}

static int goIgnoreTagWithType(gzFile file, unsigned char type) {
  if (type == 0) {
    return ERR_EARLY_TAG_END;
  } else if (type == 1) {
    gzgetc(file);
    return 0;
  } else if (type == 2) {
    gzseek(file, 2, SEEK_CUR);
  } else if (type == 3) {
    gzseek(file, 4, SEEK_CUR);
  } else if (type == 4) {
    gzseek(file, 8, SEEK_CUR);
  } else if (type == 5) {
    gzseek(file, sizeof(float), SEEK_CUR);
  } else if (type == 6) {
    gzseek(file, sizeof(double), SEEK_CUR);
  } else if (type == 7) {
    uint32_t len;
    gzread(file, &len, 4);
    len = htonl(len);
    gzseek(file, len, SEEK_CUR);
  } else if (type == 8) {
    uint16_t len;
    gzread(file, &len, 2);
    len = htons(len);
    gzseek(file, len, SEEK_CUR);
  } else if (type == 9) {
    int subtype = gzgetc(file);
    if (subtype == EOF) {
      return ERR_INVALID_SCHEMATIC_FILE;
    }
    uint32_t len;
    gzread(file, &len, 4);
    len = htonl(len);
    for (unsigned int i = 0; i < len; i++) {
      goIgnoreTagWithType(file, subtype);
    }
  } else if (type == 10) {
    while (1) {
      int subtype = gzgetc(file);
      if (subtype == 0) {
        break;
      } else if (subtype == EOF) {
        return ERR_INVALID_SCHEMATIC_FILE;
      }
      goIgnoreTagName(file);
      goIgnoreTagWithType(file, subtype);
    }
  } else if (type == 11) {
    uint32_t len;
    gzread(file, &len, 4);
    len = htonl(len);
    gzseek(file, 4 * len, SEEK_CUR);
  } else if (type == 12) {
    uint32_t len;
    gzread(file, &len, 4);
    len = htonl(len);
    gzseek(file, 4 * len, SEEK_CUR);
  } else {
    return ERR_UNEXPECTED_TAG_TYPE;
  }
  return 0;
}

static int seekToNextTag(const char *name, unsigned char type, gzFile file,
                         int origin) {
  if (origin == EOF) {
    origin = gztell(file);
  } else if (origin == gztell(file)) {
    return ERR_NOT_FOUND;
  }
  if (gztell(file) == 0) {
    if (gzgetc(file) != 10) {
      return ERR_INVALID_SCHEMATIC_FILE;
    }
    goIgnoreTagName(file);
  }
  int currentTag = gzgetc(file);
  printf("TAG %d\n", currentTag);
  if (currentTag == 0) {
    gzseek(file, 0, SEEK_SET);
    return seekToNextTag(name, type, file, origin);
    // return ERR_NOT_FOUND;
  }
  if (currentTag != type) {
    goIgnoreTagName(file);
    int err = goIgnoreTagWithType(file, currentTag);
    if (err != 0) {
      return err;
    }
    return seekToNextTag(name, type, file, origin);
  }
  uint16_t strLen;
  gzread(file, &strLen, 2);
  strLen = htons(strLen);
  if (strLen != strlen(name)) {
    gzseek(file, strLen, SEEK_CUR);
    int err = goIgnoreTagWithType(file, currentTag);
    if (err != 0) {
      return err;
    }
    return seekToNextTag(name, type, file, origin);
  }
  char *currentTagName = malloc(strLen);
  gzread(file, currentTagName, strLen);
  if (memcmp(currentTagName, name, strLen) == 0) {
    free(currentTagName);
    return 0;
  }
  free(currentTagName);
  {
    int err = goIgnoreTagWithType(file, currentTag);
    if (err != 0) {
      return err;
    }
  }
  return seekToNextTag(name, type, file, origin);
}

static int readRequiredTags(gzFile file, int origin, unsigned short **storage,
                            unsigned int nokoru) {
  if (nokoru == EOF) {
    nokoru = 6;
  } else if (nokoru == 0) {
    return 0;
  }
  if (origin == EOF) {
    origin = gztell(file);
  } else if (origin == gztell(file)) {
    if (nokoru == 3) {
      **(storage + 3) = 0;
      **(storage + 4) = 0;
      **(storage + 5) = 0;
      return 0;
    }
    return ERR_INVALID_SCHEMATIC_FILE;
  }
  if (gztell(file) == 0) {
    if (gzgetc(file) != 10) {
      return ERR_INVALID_SCHEMATIC_FILE;
    }
    goIgnoreTagName(file);
  }
  int currentTag = gzgetc(file);
  printf("[R] TAG %d\n", currentTag);
  if (currentTag == 0) {
    gzseek(file, 0, SEEK_SET);
    return readRequiredTags(file, origin, storage, nokoru);
  }
  if (currentTag != 2) {
    goIgnoreTagName(file);
    int err = goIgnoreTagWithType(file, currentTag);
    if (err != 0) {
      return err;
    }
    return readRequiredTags(file, origin, storage, nokoru);
  }
  uint16_t strLen;
  gzread(file, &strLen, 2);
  strLen = htons(strLen);
  if (strLen != 9 && strLen != 6 && strLen != 5) {
    gzseek(file, strLen, SEEK_CUR);
    int err = goIgnoreTagWithType(file, currentTag);
    if (err != 0) {
      return err;
    }
    return readRequiredTags(file, origin, storage, nokoru);
  }
  char *currentTagName = malloc(strLen);
  gzread(file, currentTagName, strLen);
  if (strLen == 9) {
    if (memcmp(currentTagName, "WEOffsetX", strLen) == 0) {
      free(currentTagName);
      gzread(file, *(storage + 3), 2);
      **(storage + 3) = htons(**(storage + 3));
      return readRequiredTags(file, origin, storage, nokoru - 1);
    } else if (memcmp(currentTagName, "WEOffsetY", strLen) == 0) {
      free(currentTagName);
      gzread(file, *(storage + 4), 2);
      **(storage + 4) = htons(**(storage + 4));
      return readRequiredTags(file, origin, storage, nokoru - 1);
    } else if (memcmp(currentTagName, "WEOffsetZ", strLen) == 0) {
      free(currentTagName);
      gzread(file, *(storage + 5), 2);
      **(storage + 5) = htons(**(storage + 5));
      return readRequiredTags(file, origin, storage, nokoru - 1);
    }
  } else if (strLen == 6) {
    if (memcmp(currentTagName, "Length", strLen) == 0) {
      free(currentTagName);
      gzread(file, *(storage + 1), 2);
      **(storage + 1) = htons(**(storage + 1));
      return readRequiredTags(file, origin, storage, nokoru - 1);
    } else if (memcmp(currentTagName, "Height", strLen) == 0) {
      free(currentTagName);
      gzread(file, *(storage + 2), 2);
      **(storage + 2) = htons(**(storage + 2));
      return readRequiredTags(file, origin, storage, nokoru - 1);
    }
  } else {
    if (memcmp(currentTagName, "Width", strLen) == 0) {
      free(currentTagName);
      gzread(file, *storage, 2);
      **storage = htons(**storage);
      return readRequiredTags(file, origin, storage, nokoru - 1);
    }
  }
  free(currentTagName);
  {
    int err = goIgnoreTagWithType(file, currentTag);
    if (err != 0) {
      return err;
    }
  }
  return readRequiredTags(file, origin, storage, nokoru);
}

extern void builder_schematic_channel_input(uint32_t channelID, int64_t x,
                                            int64_t y, int64_t z,
                                            unsigned char id,
                                            unsigned char data);

unsigned char builder_schematic_process_schematic_file(uint32_t channelID,
                                                       char *path,
                                                       int64_t beginX,
                                                       int64_t beginY,
                                                       int64_t beginZ) {
  gzFile schematic_file = gzopen(path, "rb");
  if (!schematic_file) {
    free(path);
    return ERR_INVALID_SCHEMATIC_FILE;
  }
  gzFile data_file = gzopen(path, "rb");
  char copied_path[1024];
  sprintf(copied_path, "%s", path);
  free(path);
  gzFile file = schematic_file;
  uint16_t width;
  uint16_t length;
  uint16_t height;
  uint16_t offset_x;
  uint16_t offset_y;
  uint16_t offset_z;
  unsigned short *storage[6] = {&width,    &length,   &height,
                                &offset_x, &offset_y, &offset_z};
  int err = readRequiredTags(file, EOF, storage, EOF);
  printf("Width: %d\nLength: %d\nHeight: %d\n\nOffsetX: %d\nOffsetY: "
         "%d\nOffsetZ: %d\n",
         width, length, height, offset_x, offset_y, offset_z);
  err = seekToNextTag("Blocks", 7, file, EOF);
  if (err) {
    gzclose(file);
    gzclose(data_file);
    return ERR_INVALID_SCHEMATIC_FILE;
  }
  err = seekToNextTag("Data", 7, data_file, EOF);
  if (err) {
    gzclose(file);
    gzclose(data_file);
    return ERR_INVALID_SCHEMATIC_FILE;
  }
  uint32_t blocksCount;
  gzread(file, &blocksCount, 4);
  blocksCount = htonl(blocksCount);
  gzseek(data_file, 4, SEEK_CUR);
  unsigned int i = 0;
  unsigned int toSeek = 0;
  gzFile blockSources[16];
  gzFile dataSources[16];
  blockSources[0] = file;
  dataSources[0] = data_file;
  z_off_t blocksBeginning = gztell(file);
  z_off_t dataBeginning = gztell(data_file);
  for (int i = 1; i < 16; i++) {
    blockSources[i] = gzopen(copied_path, "rb");
    dataSources[i] = gzopen(copied_path, "rb");
    gzseek(blockSources[i], blocksBeginning + width * i, SEEK_SET);
    gzseek(dataSources[i], dataBeginning + width * i, SEEK_SET);
  }
  unsigned int chunksLength = length / 16;
  if (length % 16 != 0)
    chunksLength++;
  unsigned int chunksWidth = width / 16;
  if (width % 16 != 0)
    chunksWidth++;
  unsigned int realZ = 0;
  unsigned int realX = 0;
  for (unsigned int baseX = 0; baseX <= width / 16; baseX += 16) {
    realX = baseX;
    for (unsigned int chunkZ = 0; chunkZ < chunksLength; chunkZ++) {
      realZ = chunkZ * 16;
      for (unsigned int y = 0; y < height; y++) {
        for (unsigned int cz = 0; cz < 16; cz++) {
          if (realZ >= length)
            break;
          if (chunkZ != 0) {
            gzseek(blockSources[cz],
                   blocksBeginning + y * width * length + width * chunkZ * 16 +
                       cz * width,
                   SEEK_SET);
            gzseek(dataSources[cz],
                   dataBeginning + y * width * length + width * chunkZ * 16 +
                       cz * width,
                   SEEK_SET);
          }
          for (unsigned cx = 0; cx < 16; cx++) {
            if (realX >= width)
              break;
            int currentBlock = gzgetc(blockSources[cz]);
            if (!currentBlock) {
              gzseek(dataSources[cz], 1, SEEK_CUR);
              continue;
            } else if (currentBlock == EOF) {
              // TODO: Close them !
              printf("ERR - EOF\n");
              return ERR_INVALID_SCHEMATIC_FILE;
            }
            int currentData = gzgetc(dataSources[cz]);
            if (currentData == EOF) {
              printf("ERR - EOF\n");
              return ERR_INVALID_SCHEMATIC_FILE;
            }
            builder_schematic_channel_input(
                channelID, realX + offset_x + beginX, y + offset_y + beginY,
                realZ + offset_z + beginZ, currentBlock, currentData);
            realX++;
          }
          realZ++;
          realX = baseX;
        }
        realZ = chunkZ * 16;
      }
    }
  }
  for (int i = 0; i < 16; i++) {
    gzclose(blockSources[i]);
    gzclose(dataSources[i]);
  }
  return 0;
}