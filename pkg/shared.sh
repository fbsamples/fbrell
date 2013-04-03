build() {
  unset GIT_DIR # this can interefere with "go get"

  goimport=github.com/daaku/rell
  bindir=$pkgdir/usr/bin
  install -d $bindir
  GO_LDFLAGS="-X github.com/daaku/rell/context/viewcontext.version $pkgver"

  msg "Building"
  go build -ldflags="$GO_LDFLAGS" -o=$bindir/$pkgname $goimport

  gitabs=${srcdir}/../../..

  msg "Getting go dependenices"
  (cd $gitabs && go get -v)

  msg "Getting npm dependencies"
  (cd $gitabs/js && npm install)

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
