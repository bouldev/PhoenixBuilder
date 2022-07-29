#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <zlib.h>
#include <arpa/inet.h>
#include <string.h>
#include <errno.h>

#define ERR_INVALID_SCHEMATIC_FILE 1
#define ERR_NOT_IMPLEMENTED 2
#define ERR_EARLY_TAG_END 3
#define ERR_UNEXPECTED_TAG_TYPE 4
#define ERR_NOT_FOUND 5

static void goIgnoreTagName(gzFile file) {
	uint16_t strLen;
	gzread(file, &strLen, 2);
	strLen=htons(strLen);
	//printf("%d\n",strLen);
	gzseek(file, strLen, SEEK_CUR);
}

static int goIgnoreTagWithType(gzFile file, unsigned char type) {
	if(type==0) {
		return ERR_EARLY_TAG_END;
	}else if(type==1) {
		gzgetc(file);
		return 0;
	}else if(type==2) {
		gzseek(file, 2, SEEK_CUR);
	}else if(type==3) {
		gzseek(file, 4, SEEK_CUR);
	}else if(type==4) {
		gzseek(file, 8, SEEK_CUR);
	}else if(type==5) {
		gzseek(file, sizeof(float), SEEK_CUR);
	}else if(type==6) {
		gzseek(file, sizeof(double), SEEK_CUR);
	}else if(type==7) {
		uint32_t len;
		gzread(file, &len, 4);
		len=htonl(len);
		gzseek(file, len, SEEK_CUR);
	}else if(type==8) {
		uint16_t len;
		gzread(file, &len, 2);
		len=htons(len);
		gzseek(file, len, SEEK_CUR);
	}else if(type==9) {
		int subtype=gzgetc(file);
		if(subtype==EOF) {
			return ERR_INVALID_SCHEMATIC_FILE;
		}
		uint32_t len;
		gzread(file, &len, 4);
		len=htonl(len);
		for(unsigned int i=0;i<len;i++) {
			goIgnoreTagWithType(file, subtype);
		}
	}else if(type==10) {
		while(1) {
			int subtype=gzgetc(file);
			if(subtype==0) {
				break;
			}else if(subtype==EOF) {
				return ERR_INVALID_SCHEMATIC_FILE;
			}
			goIgnoreTagName(file);
			goIgnoreTagWithType(file, subtype);
		}
	}else if(type==11) {
		uint32_t len;
		gzread(file, &len, 4);
		len=htonl(len);
		gzseek(file, 4*len, SEEK_CUR);
	}else if(type==12) {
		uint32_t len;
		gzread(file, &len, 4);
		len=htonl(len);
		gzseek(file, 4*len, SEEK_CUR);
	}else{
		return ERR_UNEXPECTED_TAG_TYPE;
	}
	return 0;
}

static int seekToNextTag(const char *name, unsigned char type, gzFile file, int origin) {
	if(origin==EOF) {
		origin=gztell(file);
	}else if(origin==gztell(file)) {
		return ERR_NOT_FOUND;
	}
	if(gztell(file)==0) {
		if(gzgetc(file)!=10) {
			return ERR_INVALID_SCHEMATIC_FILE;
		}
		goIgnoreTagName(file);
	}
	int currentTag=gzgetc(file);
	printf("TAG %d\n",currentTag);
	if(currentTag==0) {
		gzseek(file, 0, SEEK_SET);
		return seekToNextTag(name, type, file, origin);
		//return ERR_NOT_FOUND;
	}
	if(currentTag!=type) {
		goIgnoreTagName(file);
		int err=goIgnoreTagWithType(file, currentTag);
		if(err!=0) {
			return err;
		}
		return seekToNextTag(name, type, file, origin);
	}
	uint16_t strLen;
	gzread(file, &strLen, 2);
	strLen=htons(strLen);
	if(strLen!=strlen(name)) {
		gzseek(file, strLen, SEEK_CUR);
		int err=goIgnoreTagWithType(file, currentTag);
		if(err!=0) {
			return err;
		}
		return seekToNextTag(name, type, file, origin);
	}
	char *currentTagName=malloc(strLen);
	gzread(file, currentTagName, strLen);
	if(memcmp(currentTagName, name, strLen)==0) {
		free(currentTagName);
		return 0;
	}
	free(currentTagName);
	{
		int err=goIgnoreTagWithType(file, currentTag);
		if(err!=0) {
			return err;
		}
	}
	return seekToNextTag(name, type, file, origin);
}

extern void builder_schematic_channel_input(uint32_t channelID, int64_t x, int64_t y, int64_t z, unsigned char id, unsigned char data);

unsigned char builder_schematic_process_schematic_file(uint32_t channelID, char *path, int64_t beginX, int64_t beginY, int64_t beginZ) {
	gzFile schematic_file=gzopen(path, "rb");
	if(!schematic_file) {
		free(path);
		return ERR_INVALID_SCHEMATIC_FILE;
	}
	gzFile data_file=gzopen(path, "rb");
	free(path);
	gzFile file=schematic_file;
	uint16_t width;
	uint16_t length;
	uint16_t height;
	uint16_t offset_x;
	uint16_t offset_y;
	uint16_t offset_z;
	int err=seekToNextTag("Width", 2, file, EOF);
	if(err) {
		gzclose(file);
		gzclose(data_file);
		return err;
	}
	gzread(file, &width, 2);
	err=seekToNextTag("Length", 2, file, EOF);
	if(err) {
		gzclose(file);
		gzclose(data_file);
		return err;
	}
	gzread(file, &length, 2);
	err=seekToNextTag("Height", 2, file, EOF);
	if(err) {
		gzclose(file);
		gzclose(data_file);
		return err;
	}
	gzread(file, &height, 2);
	width=htons(width);
	length=htons(length);
	height=htons(height);
	err=seekToNextTag("WEOffsetX", 2, file, EOF);
	if(!err) {
		gzread(file, &offset_x, 2);
		err=seekToNextTag("WEOffsetY", 2, file, EOF);
		if(err) {
			gzclose(file);
			gzclose(data_file);
			return ERR_INVALID_SCHEMATIC_FILE;
		}
		gzread(file, &offset_y, 2);
		err=seekToNextTag("WEOffsetZ", 2, file, EOF);
		if(err) {
			gzclose(file);
			gzclose(data_file);
			return ERR_INVALID_SCHEMATIC_FILE;
		}
		gzread(file, &offset_z, 2);
		offset_x=htons(offset_x);
		offset_y=htons(offset_y);
		offset_z=htons(offset_z);
	}else{
		printf("TRACE >> No WEOffset*\n");
		offset_x=0;
		offset_y=0;
		offset_z=0;
	}
	printf("Width: %d\nLength: %d\nHeight: %d\n\nOffsetX: %d\nOffsetY: %d\nOffsetZ: %d\n",
		width,length,height,offset_x,offset_y,offset_z);
	err=seekToNextTag("Blocks", 7, file, EOF);
	if(err) {
		gzclose(file);
		gzclose(data_file);
		return ERR_INVALID_SCHEMATIC_FILE;
	}
	err=seekToNextTag("Data", 7, data_file, EOF);
	if(err) {
		gzclose(file);
		gzclose(data_file);
		return ERR_INVALID_SCHEMATIC_FILE;
	}
	uint32_t blocksCount;
	gzread(file, &blocksCount, 4);
	blocksCount=htonl(blocksCount);
	gzseek(data_file, 4, SEEK_CUR);
	unsigned int i=0;
	unsigned int toSeek=0;
	for(unsigned int y=0;y<height;y++) {
		for(unsigned int z=0;z<length;z++) {
			for(unsigned int x=0;x<width;x++) {
				int currentBlock=gzgetc(file);
				if(currentBlock==EOF) {
					printf("H %s\n",gzerror(data_file, &i));
					gzclose(file);
					gzclose(data_file);
					return ERR_INVALID_SCHEMATIC_FILE;
				}
				if(!currentBlock) {
					// currentBlock == 0 (air)
					toSeek++;
					continue;
				}
				if(toSeek) {
					gzseek(data_file, toSeek, SEEK_CUR);
					toSeek=0;
				}
				int currentData=gzgetc(data_file);
				if(currentData==EOF) {
					gzclose(file);
					gzclose(data_file);
					return ERR_INVALID_SCHEMATIC_FILE;
				}
				builder_schematic_channel_input(channelID, x+offset_x+beginX, y+offset_y+beginY, z+offset_z+beginZ, currentBlock, currentData);
			}
		}
	}
	gzclose(file);
	gzclose(data_file);
	return 0;
}