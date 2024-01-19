export type ClaimRequest = {
  address: string
  rollupName: string
}


// returns json data from response if response is ok, otherwise throws an error
export async function getResData(response: Response): Promise<any> {
  if (response.ok) {
    return response.json()
  }
  const data = await response.json()
  throw new Error(data.message)
}
