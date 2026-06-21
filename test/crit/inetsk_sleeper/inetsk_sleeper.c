#define _GNU_SOURCE
#include <arpa/inet.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <unistd.h>

/*
 * Bind a TCP listen socket and sleep. Used to produce files.img entries
 * with INETSK for crit humanize integration tests.
 */
int main(void)
{
	int fd;
	struct sockaddr_in addr = { 0 };

	fd = socket(AF_INET, SOCK_STREAM, 0);
	if (fd < 0) {
		perror("socket");
		return 1;
	}

	addr.sin_family = AF_INET;
	addr.sin_addr.s_addr = htonl(INADDR_ANY);
	addr.sin_port = htons(0);

	if (bind(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
		perror("bind");
		return 1;
	}
	if (listen(fd, 1) < 0) {
		perror("listen");
		return 1;
	}

	printf("%d\n", getpid());
	fflush(stdout);

	while (1)
		sleep(3600);

	return 0;
}
