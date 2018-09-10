# $FreeBSD$

PORTNAME=	spirit
DISTVERSION=	master
CATEGORIES=	sysutils

MAINTAINER=	adam@mediadrive.ca
COMMENT=	Manage your daemons

LICENSE=	MIT

ONLY_FOR_ARCHS=	amd64 i386

USES=		go

USE_GITHUB=	yes
GH_ACCOUNT=	adam12
GH_PROJECT=	spirit
GH_TUPLE=	direnv:go-dotenv:4738ec4:godotenv/src/github.com/direnv/go-dotenv \
		hecticjeff:procfile:6779278:procfile/src/github.com/hecticjeff/procfile
GH_SUBDIR=	src/github.com/${GH_ACCOUNT_DEFAULT}/${PORTNAME}

GO_PKGNAME=	spirit


do-build:
	@cd ${WRKSRC}/${GH_SUBDIR} && ${SETENV} ${MAKE_ENV} GOPATH=${WRKSRC} go build

do-install:
	@${INSTALL_PROGRAM} ${WRKSRC}/${GH_SUBDIR}/${GO_PKGNAME} ${STAGEDIR}${PREFIX}/bin

.include <bsd.port.mk>
