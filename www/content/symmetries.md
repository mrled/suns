+++
title = "symmetries"
+++

A domain name, or a pair of domain names, might be symmetrical in one of several ways.

## Supported symmetries

<table>
  <thead>
    <tr>
      <td>Type</td>
      <td>Example</td>
      <td>Supported</td>
      <td>Notes</td>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Single Palindrome</td>
      <td><code>zb.snus.suns.bz</code></td>
      <td>Yes</td>
      <td><code>zb.snus</code> is <code>suns.bz</code> with letters in reverse order</td>
    </tr>
    <tr>
      <td>Double Palindrome</td>
      <td>
        <code>su.suns.bz</code>, <code>zb.snus.us</code>
        <br/>(example domain we don't own)
      </td>
      <td>Yes</td>
      <td><code>su.suns.bz</code> is <code>zb.snus.us</code> with letters in reverse order</td>
    </tr>
    <tr>
      <td>Single 180° Flip</td>
      <td><code>zq.suns.bz</code></td>
      <td>Yes</td>
      <td><code>zq.su</code> is <code>ns.bz</code> flipped 180°</td>
    </tr>
    <tr>
      <td>Double 180° Flip</td>
      <td>
        <code>zq.su</code>, <code>ns.bz</code>
        <br/>(example domains we don't own)
      </td>
      <td>Yes</td>
      <td><code>zq.su</code> is <code>ns.bz</code> flipped 180°</td>
    </tr>
    <tr>
      <td>Single Mirrored DNS Components</td>
      <td>
        <code>com.example.www.example.com</code>
      </td>
      <td>Work in progress</td>
      <td>Each domain name component is reversed</td>
    </tr>
    <tr>
      <td>Double Mirrored DNS Components</td>
      <td>
        <code>me.example.com</code>, <code>com.example.me</code>
      </td>
      <td>Yes</td>
      <td>Each domain name component is reversed</td>
    </tr>
    <tr>
      <td>Mirrored Text</td>
      <td>
        <code>duq.xodbox.pub</code>
        <br/>(example domain we don't own)
      </td>
      <td>Work in progress</td>
      <td><code>duq.xod</code> is <code>box.pub</code> as read in a mirror</td>
    </tr>
    <tr>
      <td>Double Mirrored Text</td>
      <td>
        <code>ood.pub</code>, <code>duq.boo</code>
        <br/>(example domains we don't own)
      </td>
      <td>Work in progress</td>
      <td><code>ood.pub</code> is <code>duq.boo</code> as read in a mirror</td>
    </tr>
  </tbody>
</table>

## Bonuses

We don't do anything special with symmetries like this,
but they're neat.

<table>
  <thead>
    <tr>
      <td>Type</td>
      <td>Example</td>
      <td>Supported</td>
      <td>Notes</td>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Palindrome URL</td>
      <td><code>https://zb.snus.suns.bz//:sptth</code></td>
      <td>No</td>
      <td>
        <code>https://zb.snus</code> is the reverse of <code>suns.bz//:sptth</code>
      </td>
    </tr>
    <tr>
      <td>Palindrome Email</td>
      <td><code>zb.snus@suns.bz</code></td>
      <td>No</td>
      <td>
        <code>zb.snus</code> is the reverse of <code>suns.bz</code>
      </td>
    </tr>
    <tr>
      <td>180° Flip URL</td>
      <td><code>https://zq.suns.bz//:sdʇʇɥ</code></td>
      <td>No</td>
      <td>
        <code>https://zq.su</code> is <code>ns.bz//:sdʇʇɥ</code> flipped 180°
        <br/>(note non-ASCII <code>:sdʇʇɥ</code>)
      </td>
    </tr>
    <tr>
      <td>180° Flip Email</td>
      <td><code>zq.suns@suns.bz</code></td>
      <td>No</td>
      <td>
        <code>zq.suns</code> is <code>suns.bz</code> flipped 180°
      </td>
    </tr>
    <tr>
      <td>Mirrored Text Email</td>
      <td><code>moc.elpmaxe@example.com</code></td>
      <td>No</td>
      <td>
        <code>moc.elpmaxe</code> is <code>example.com</code> as read in a mirror
      </td>
    </tr>
    <tr>
      <td>Antonymic DNS</td>
      <td><code>https://at.example.email</code>, <code>https@example.website</code></td>
      <td>No</td>
      <td>
        Inspired by <code>https://slashdot.org</code>
      </td>
    </tr>
  </tbody>
</table>

Some mirrored or flipped symmetries are possible with non-ASCII characters,
but do not pass IDNA character validation.
For example, `ɯoɔ.ǝldɯɐxǝ.example.com` would be a neat flip,
and the browser will accept that as a domain name,
but it will translate it in the URL bar into punycode: `https://xn--o-10a3f.xn--ldx-5ebd20eyg.example.com/`,
which isn't as fun.
Some of these non-ASCII characters *are* valid,
but finding them can be tricky.
To avoid homograph attacks,
where characters look the same but are not,
browsers have complicated rules for when characters are shown as Unicode or as punycode.
We encourage you tp spend as much time researching this as possible.

If you're looking for more bonus points,
you might try to get the shortest possible flip as measured in either characters or domain name components.
At the time of this writing, for example, the valid ASCII 180° flip `zznq.buzz` is available!
Sure, it's almost six thousand dollars,
but can you really put a price on winning a fake game on the Internet?
