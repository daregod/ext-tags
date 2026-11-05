class ExtTags < Formula
  desc "Write and import foobar2000 external .tag files"
  homepage "https://github.com/daregod/ext-tags"
  version "0.10.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.3/ext-tags_0.10.3_darwin_arm64.tar.gz"
      sha256 "491ea5a39fb3cc830b0ac633b4c4aa5eb2a5725fd2ffc76a8fa7b236c2e2c92b"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.3/ext-tags_0.10.3_darwin_amd64.tar.gz"
      sha256 "32927fdf0f1cef015d76b35ca0effc32031bbdbdfd34a045b768868657314aed"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.3/ext-tags_0.10.3_linux_arm64.tar.gz"
      sha256 "d8f92ac65994e99b9e7b4ccdabcf48f5d76e86f7b9825955d5c4ce174500fb9c"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.3/ext-tags_0.10.3_linux_amd64.tar.gz"
      sha256 "b5ae335d24134e023e78887c7e26169f617850ce3f8eac8baddee2e1518bc44b"
    end
  end

  def install
    binary = Dir["ext-tags_*"].first
    odie "ext-tags binary not found" if binary.nil?
    bin.install binary => "ext-tags"
  end

  test do
    assert_match "ext-tags", shell_output("#{bin}/ext-tags version")
  end
end
