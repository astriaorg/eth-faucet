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
        <div class="hero-body">
            <div class="container has-text-centered">
                <div class="column is-6 is-offset-3">
                    {#each rollups as rollup}
                        <div class="card p-5">
                            <header class="card-header">
                                <p class="card-header-title is-size-5 has-text-weight-normal has-text-light">
                                    {rollup.name}
                                </p>
                            </header>
                            <div class="card-content p-4">
                                <button class="button is-ghost is-outlined-light">
                                    <a href="/{rollup.name}">Get Testnet Tokens</a>
                                </button>
                            </div>
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
</style>
