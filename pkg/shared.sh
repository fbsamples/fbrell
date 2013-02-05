build() {
  unset GIT_DIR # this can interefere with "go get"
  msg "Cleaning old build"
  rm -rf $pkgdir/*

  export GOPATH=${srcdir}
  cd ${GOPATH}
  goimport=github.com/daaku/rell
  gitabs=${GOPATH}/src/$goimport

  mkdir -p $(dirname ${gitabs})
  cd $(dirname ${gitabs})
  rsync \
    --archive \
    --one-file-system \
    --sparse \
    --quiet \
    --delete \
    --exclude=pkg \
    --exclude=.git \
    $srcdir/../../../  $(basename ${gitabs})/

  cd $gitabs

  msg "Getting go dependenices"
  go get -v

  msg "Getting npm dependencies"
  (cd js && npm install)

  bindir=$pkgdir/usr/bin
  mkdir -p $bindir
  binfile=$bindir/$pkgname
  msg "Building"
  GO_LDFLAGS="-X github.com/daaku/rell/context/viewcontext.version $pkgver"
  go build -ldflags="$GO_LDFLAGS" -o=$binfile $goimport

  msg "Copying resources"
  install -d $gitabs/public $pkgdir/usr/share/$pkgname/public
  cp -r $gitabs/public $pkgdir/usr/share/$pkgname
  install -d $gitabs/examples/db/mu $pkgdir/usr/share/$pkgname/examples/mu
  cp -r $gitabs/examples/db/mu $pkgdir/usr/share/$pkgname/examples
  install -d $gitabs/examples/db/old $pkgdir/usr/share/$pkgname/examples/old
  cp -r $gitabs/examples/db/old $pkgdir/usr/share/$pkgname/examples

  msg "Installing systemd service & socket"
  install -D $srcdir/../$pkgname.service $pkgdir/usr/lib/systemd/system/$pkgname.service
  install -D $srcdir/../$pkgname.socket $pkgdir/usr/lib/systemd/system/$pkgname.socket

  msg "Creating static resources"
  cd $gitabs/js
  ./node_modules/.bin/browserify -e rell.js --exports require > $pkgdir/usr/share/$pkgname/browserify.js
}
