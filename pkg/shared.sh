build() {
  unset GIT_DIR
  msg "Cleaning old build"
  rm -rf $pkgdir/*

  export GOPATH=${srcdir}
  cd ${GOPATH}
  gitabs=${GOPATH}/${_goimport}

  if [ -d ${_goimport} ]; then
    msg "Updating existing repository"
    cd ${_goimport}
    git pull
  else
    msg "Initial clone"
    mkdir -p ${_gitcontainer}
    cd ${_gitcontainer}
    git clone ${_gitroot}
  fi

  cd $gitabs

  msg "Getting go dependenices"
  go get -v

  msg "Getting npm dependencies"
  (cd public && npm install)

  bindir=$pkgdir/usr/bin
  mkdir -p $bindir
  binfile=$bindir/$pkgname
  msg "Building"
  go build -o $binfile

  msg "Copying resources"
  install -d $gitabs/public $pkgdir/usr/share/$pkgname/public
  cp -r $gitabs/public $pkgdir/usr/share/$pkgname
  install -d $gitabs/examples/db/mu $pkgdir/usr/share/$pkgname/examples/mu
  cp -r $gitabs/examples/db/mu $pkgdir/usr/share/$pkgname/examples
  install -d $gitabs/examples/db/old $pkgdir/usr/share/$pkgname/examples/old
  cp -r $gitabs/examples/db/old $pkgdir/usr/share/$pkgname/examples

  msg "Creating rc script"
  rcname=$pkgdir/etc/rc.d/$pkgname
  mkdir -p $(dirname $rcname)
  rc > $rcname
  chmod +x $rcname

  msg "Creating static resources"
  cd $gitabs/public
  ./node_modules/.bin/browserify -e rell.js > $pkgdir/usr/share/$pkgname/browserify.js
}

rc() {
  cat $srcdir/../../rc | sed -e "s/TOKEN_PKGNAME/$pkgname/g"
}
