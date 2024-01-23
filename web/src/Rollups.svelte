<script lang="ts" context="module">
  import { setDefaults as setToast } from 'bulma-toast'

  // config
  document.title = `RIA Faucet`
  setToast({
    position: 'bottom-center',
    dismissible: true,
    pauseOnHover: true,
    closeOnClick: false,
    animate: { in: 'fadeIn', out: 'fadeOut' },
  })

  type Rollup = {
    name: string,
    chainId: number,
  }

  type Rollups = Rollup[]

</script>

<script lang="ts">
  import { onMount } from 'svelte'

  let rollups = []

  onMount(async () => {
    console.log('on mount')
    try {
      const res = await fetch('/api/rollups')
      rollups = await res.json()
    } catch (e) {
      console.error(e)
    }
  })


</script>

<main>
    <section class="hero is-info is-fullheight">
        <div class="hero-head">
            <nav class="navbar">
                <div class="container">
                    <div class="navbar-brand">
                        <a
                                class="navbar-item is-white"
                                href="https://astria.org"
                                target="_blank"
                        >
                            <span><b>Astria</b></span>
                        </a>
                    </div>

                    <div id="navbarMenu" class="navbar-menu">
                        <div class="navbar-end">
              <span class="navbar-item">
                <a
                        class="button is-white is-outlined"
                        href="https://github.com/astriaorg/eth-faucet"
                        target="_blank"
                >
                  <span class="icon">
                    <i class="fa fa-github"/>
                  </span>
                  <span>View Source</span>
                </a>
              </span>
                        </div>
                    </div>
                </div>
            </nav>
        </div>

        <div class="hero-body">
            <div class="container has-text-centered">
                <div class="column is-6 is-offset-3">
                    {#each rollups as rollup}
                        <div class="box">
                            <h1 class="title">
                                {rollup.name}
                            </h1>
                            <h2 class="subtitle">
                                <a href="/faucet/{rollup.chainId}">Get Testnet Tokens</a>
                            </h2>
                        </div>
                    {/each}
                </div>
            </div>
        </div>
    </section>
</main>

<style>
    .hero.is-info {
        background: black url('/hero-blocks.webp') no-repeat fixed center center;
        -webkit-background-size: cover;
        -moz-background-size: cover;
        -o-background-size: cover;
        background-size: cover;
    }

    .hero.is-info a.navbar-item:hover {
        background-color: transparent;
    }

    .hero .subtitle {
        padding: 3rem 0;
        line-height: 1.5;
    }

    .box {
        border-radius: 0;
        background: transparent;
    }

    .button {
        border-radius: 0;
    }
</style>
