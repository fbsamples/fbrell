<h1>Using the Graph API for Page Admin</h1>

<p>If you are the admin of a Facebook Page, the Graph API can be used to control that Page, update it, and publish posts to it. The example below shows some simple code that can be used for this. Note, this example will use your live Facebook Pages, so only use it with any that you're comfortable publishing a test post to.</p>

<hr />

<h2>Getting your Page Access Token</h2>

<p>First, we'll need the <code>manage_pages</code> permission to see the pages you admin, so we'll insert a Login button which requests the correct permissions (click on this if you haven't already granted the permission):</p>

<div class="fb-login-button" data-scope="manage_pages" data-max-rows="1" data-size="medium"></div>

<hr />

<h2>Requesting the Pages</h2>

<div id="pageBtn" class="btn btn-success clearfix">Click me to show the list of Pages you admin.</div>

<ul id="pagesList" class="btn-group btn-group-vertical clearfix"></ul>

<script>
document.getElementById('pageBtn').onclick = function() {
  FB.api('/me/accounts?fields=name,access_token,link', function(response) {
    Log.info('API response', response);
    var list = document.getElementById('pagesList');
    for (var i=0; i < response.data.length; i++) {
      var li = document.createElement('li');
      li.innerHTML = response.data[i].name;
      li.dataset.token = response.data[i].access_token;
      li.dataset.link = response.data[i].link;
      li.className = 'btn btn-mini';
      li.onclick = function() {
        document.getElementById('pageName').innerHTML = this.innerHTML;
        document.getElementById('pageToken').innerHTML = this.dataset.token;
        document.getElementById('pageLink').setAttribute('href', this.dataset.link);
      }
      list.appendChild(li);
    }
  });
  return false;
}  
</script>

<hr />

<h2>Publishing to a Page</h2>

<p>Click on a page in the list created above, and the name and Page access token will be shown here:</p>

<span id="pageName" class="label label-success">No page selected</span>
<span id="pageToken" class="label label-success">No page selected</span>

<p>Once you've chosen a Page, you can click the button below to publish a "Hello, world!" post to that Page. Be careful, this will be visible to anyone who is a fan of that Page! This button simply triggers another FB.api() publishing call, but this time the access token of the Page is specified to replace the user access token we've used before.</p>
<div id="publishBtn" class="btn btn-success">Publish me!</div>

<script>
document.getElementById('publishBtn').onclick = function() {
  var pageToken = document.getElementById('pageToken').innerHTML;
  FB.api('/me/feed', 'post', {message: 'Hello, world!', access_token: pageToken}, function(response) {
    Log.info('API response', response);
    document.getElementById('publishBtn').innerHTML = 'API response is ' + response.id;
  });
  return false;
}  
</script>

<p>Now go look at <a id="pageLink" href="#">your page</a> and if the API request was successful, you'll see the Hello, World post in the feed.</p>

<hr />

<h3>Related Guides</h3>

<p>Read <a href="https://developers.facebook.com/docs/javascript/quickstart/#graphapi">our quickstart to using the JavaScript SDK for Graph API calls</a> for more info.</p>
