class ExtTags < Formula
  desc "Write and import foobar2000 external .tag files"
  homepage "https://github.com/daregod/ext-tags"
  version "0.10.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.2/ext-tags_0.10.2_darwin_arm64.tar.gz"
      sha256 "953af252917fdf32169241d6c6d0c7a26eea93c6a35496b8ee4fd720641e562c"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.2/ext-tags_0.10.2_darwin_amd64.tar.gz"
      sha256 "9711695743489afa797476fe3b8fa660fe1386fa7f429e0d042bd6e4200ec0da"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.2/ext-tags_0.10.2_linux_arm64.tar.gz"
      sha256 "d7cc94ba28ac34400f08490c71590e45ffaa9e68b4b7ef3e9c72aa2cac8c81ed"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.2/ext-tags_0.10.2_linux_amd64.tar.gz"
      sha256 "80770081a6206accf3ad786b587eacae2b88bdcef5f6c41a1a67cb969b8e6f42"
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
