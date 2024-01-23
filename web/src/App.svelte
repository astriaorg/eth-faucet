<script>
  import 'animate.css';
  import 'bulma/css/bulma.css';
  import { Router, Link, Route } from "svelte-routing";
  import Faucet from './Faucet.svelte';
  import Rollups from './Rollups.svelte'
  export let url = "";
</script>

<svelte:head>
  <link
    href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css"
    rel="stylesheet"
  />
</svelte:head>

<Router {url}>
  <div>
    <Route path="/:rollupName"><Faucet /></Route>
    <Route path="/"><Rollups /></Route>
  </div>
</Router>
