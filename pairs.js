module.exports = Pairs

function startsWith(big, little) {
  if (big.length < little.length) return false
  if (big.length === little.length) return big === little
  for (var i=0, l=little.length; i<l; i++) {
    if (big[i] !== little[i]) return false
  }
  return true
}

function Pairs(pairs) {
  if (!(this instanceof Pairs)) return new Pairs(pairs)
  if (!Array.isArray(pairs)) throw new Error('pairs must be an Array')
  pairs.forEach(function(pair) {
    if (!Array.isArray(pair))
      throw new Error('each entry in the pairs must be an Array')
    if (pair.length !== 2)
      throw new Error(
        'each entry in the pairs must be an Array with exactly 2 elements')
  })
  this.pairs = pairs
}

Pairs.prototype.addPair = function(property, content) {
  this.pairs.push([property, content])
  return this
}

Pairs.prototype.getPairs = function() {
  return this.pairs
}

Pairs.prototype.getPairsByPrefix = function(prefix) {
  return this.pairs.filter(function(pair) {
    return startsWith(pair[0], prefix)
  })
}

Pairs.prototype.getPairsByName = function(property) {
  return this.pairs.filter(function(pair) {
    return pair[0] === property
  })
}

Pairs.prototype.hasPairWithName = function(property) {
  return this.getPairsByName(property).length > 0
}

Pairs.prototype.getFirstByName = function(property) {
  var first = this.getPairsByName(property)[0]
  return first && first[1]
}

Pairs.prototype.isEmpty = function() {
  return this.pairs.length === 0
}
