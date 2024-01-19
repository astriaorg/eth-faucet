<script lang="ts">
  import { onMount } from 'svelte'
  import { getAddress } from '@ethersproject/address'
  import { CloudflareProvider } from '@ethersproject/providers'
  import { setDefaults as setToast, toast, ToastType } from 'bulma-toast'
  import { ClaimRequest, getResData } from './api'


  type FaucetInfo = {
    fundingAddress: string
    rollupName: string
    network: string
    payout: number
  }

  let input = null
  let faucetInfo: FaucetInfo = {
    fundingAddress: '',
    rollupName: '',
    network: 'testnet',
    payout: 1,
  }

  $: document.title = `RIA Faucet`

  function getRollupNameFromPath(): string {
    let path = window.location.pathname
    if (path[0] === '/') {
      path = path.slice(1)
    }
    if (path === '') {
      throw new Error('No rollup name found in path')
    }
    const pattern = /^[a-z]+[a-z0-9]*(?:-[a-z0-9]+)*$/
    if (!pattern.test(path)) {
      throw new Error(`Invalid rollup name: ${path}`)
    }
    return path
  }

  setToast({
    position: 'bottom-center',
    dismissible: true,
    pauseOnHover: true,
    closeOnClick: false,
    animate: { in: 'fadeIn', out: 'fadeOut' },
  })

  onMount(async () => {
    if (window.location.pathname === '/') {
      // TODO - get full list of rollups?
      return
    }

    try {
      const rollupName = getRollupNameFromPath()
      const res = await fetch(`/api/info/${rollupName}`)
      faucetInfo = await getResData(res)
    } catch (error) {
      console.error(error)
      toast({ message: error.message, type: 'is-warning', position: 'center' })
      setTimeout(() => {
        window.location.href = '/'
      }, 1000)
    }

  })

  async function handleRequest() {
    let address = input
    if (address.endsWith('.eth')) {
      try {
        const provider = new CloudflareProvider()
        address = await provider.resolveName(address)
        if (!address) {
          toast({ message: 'invalid ENS name', type: 'is-warning' })
          return
        }
      } catch (error) {
        toast({ message: error.reason, type: 'is-warning' })
        return
      }
    }

    try {
      address = getAddress(address)
    } catch (error) {
      toast({ message: error.reason, type: 'is-warning' })
      return
    }

    const req: ClaimRequest = {
      address,
      rollupName: faucetInfo.rollupName,
    }
    const res = await fetch('/api/claim', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })

    let { message } = await res.json()
    let type: ToastType = res.ok ? 'is-success' : 'is-warning'
    toast({ message, type })
  }
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
                    <h1 class="title">
                        Receive {faucetInfo.payout} RIA per request
                    </h1>
                    <h2 class="subtitle is-6">
                        Serving from {faucetInfo.fundingAddress}
                    </h2>
                    <div class="box">
                        <div class="field is-grouped">
                            <p class="control is-expanded">
                                <input
                                        bind:value={input}
                                        class="input is-rounded"
                                        type="text"
                                        placeholder="Enter your address or ENS name"
                                />
                            </p>
                            <p class="control">
                                <button
                                        on:click={handleRequest}
                                        class="button is-white is-outlined"
                                >
                                    Request
                                </button>
                            </p>
                        </div>
                    </div>
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
