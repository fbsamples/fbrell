require 'rubygems'

require 'haml'
require 'json'
require 'net/http'
require 'sinatra/base'
require 'uri'

DefaultConfig = {
  'apikey'    => 'ef8f112b63adfc86f5430a1b566f4dc1',
  'build_dev' => false,
  'comps'     => '',
  'level'     => 'debug',
  'locale'    => 'en_US',
  'old_debug' => 6,
  'server'    => 'static.ak.connect',
  'trace'     => 1,
  'version'   => 'mu',
}

class Hash
  # Returns a hash that represents the difference between two hashes.
  #
  # Examples:
  #
  #   {1 => 2}.diff(1 => 2)         # => {}
  #   {1 => 2}.diff(1 => 3)         # => {1 => 2}
  #   {}.diff(1 => 2)               # => {1 => 2}
  #   {1 => 2, 3 => 4}.diff(1 => 2) # => {3 => 4}
  def diff(h2)
    dup.delete_if { |k, v| h2[k] == v }.merge!(h2.dup.delete_if { |k, v| has_key?(k) })
  end

  def to_params
    params = ''
    stack = []

    each do |k, v|
      if v.is_a?(Hash)
        stack << [k,v]
      else
        params << "#{k}=#{v}&"
      end
    end

    stack.each do |parent, hash|
      hash.each do |k, v|
        if v.is_a?(Hash)
          stack << ["#{parent}[#{k}]", v]
        else
          params << "#{parent}[#{k}]=#{v}&"
        end
      end
    end

    params.chop! # trailing &
    params
  end
end


class Rell < Sinatra::Base
  set :haml, {format: :html5}
  configure :development do
    set :static, true
    set :public, File.dirname(__FILE__) + '/public'
  end

  helpers do
    def url(path)
      qs = @config.diff(DefaultConfig).to_params
      return path if qs == ''
      path + '?' + qs
    end

    def examples
      glob = @config['version'] == 'mu' ? 'examples/*/*' : 'examples-old/*/*'
      res = {}
      Dir[glob].map do |f|
        _, category, file = f.split('/')
        name = file[0..-6] # drop .html
        res['/' + category + '/' + name] = {
          :category => category,
          :file     => f,
          :name     => name,
        }
      end
      res
    end

    def scripts
      server = @config['server']
      server = 'www.naitik.dev575.snc1' if server == 'sb'
      url = 'http://' + server + '.facebook.com/'

      if @config['version'] == 'mu'
        if %w{snc intern beta sandcastle latest dev}.any? {|a| server.include?(a) }
          if @config['build_dev']
            url += 'connect/en_US/core.js'
          else
            url += 'assets.php/' + @config['locale'] + '/all.js'
          end
        else
          if @config['build_dev']
            url += 'connect/' + @config['locale'] + '/core.debug.js'
          else
            url += 'connect/' + @config['locale'] + '/core.js'
          end
        end
      else
        url += 'connect.php/' + @config['locale'] + '/js/' + @config['comps']
      end

      [
        'http://origin.daaku.org/js-delegator/delegator.js',
        '/json2.js',
        '/jsDump-1.0.0.js',
        '/log.js',
        '/tracer.js',
        '/rell.js',
        '/codemirror/js/codemirror.js',
        url,
      ]
    end
  end

  before do
    @title = 'FB Read Eval Log Loop'
    @config = DefaultConfig.merge(params)
    @example_code = ''
  end

  not_found do
    haml :not_found
  end

  get '/:category/:name' do |category, name|
    path = '/' + category + '/' + name
    pass unless examples.has_key? path
    #TODO memoize in production
    example = examples[path]
    @example_code = File.read(example[:file])
    haml :index
  end

  get '/' do
    haml :index
  end

  get '/help' do
    haml :help
  end

  get '/examples' do
    haml :examples
  end
end
