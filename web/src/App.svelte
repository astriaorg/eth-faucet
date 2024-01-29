<script lang="ts">
  import 'animate.css'
  import 'bulma/css/bulma.css'
  import { Route, Router } from 'svelte-routing'
  import Faucet from './Faucet.svelte'
  import Rollups from './Rollups.svelte'

  import logo from './logo.svg'
  import './scss/index.scss'

  export let url = ''
</script>

<svelte:head>
    <link
            href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css"
            rel="stylesheet"
    />
</svelte:head>

<Router {url}>
    <div>
        <nav class="navbar is-spaced is-transparent is-flex"
             aria-label="main navigation">
            <div class="navbar-brand">
                <a class="navbar-item" href="/">
                    <img src={logo} width="161" height="32" alt="logo"/>
                </a>
            </div>
            <div class="navbar-end">
              <span class="navbar-item">
                <a class="button is-white is-outlined"
                   href="https://github.com/astriaorg/eth-faucet"
                   target="_blank"
                   rel="noreferrer">
                  <span class="icon">
                    <i class="fa fa-github"/>
                  </span>
                  <span>View Source</span>
                </a>
              </span>
            </div>
        </nav>
        <Route path="/:rollupName">
            <Faucet/>
        </Route>
        <Route path="/">
            <Rollups/>
        </Route>
    </div>
</Router>
