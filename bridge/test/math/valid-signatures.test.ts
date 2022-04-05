import { ethers } from "hardhat";
import { CryptoLibraryWrapper } from "../../typechain-types";
import { expect } from "../chai-setup";
import { signedData } from "./assets/4-validators-1000-snapshots";

describe("CryptoLibrary: Validate Signature", () => {
  let crypto: CryptoLibraryWrapper;

  before(async () => {
    crypto = await (
      await ethers.getContractFactory("CryptoLibraryWrapper")
    ).deploy();
  });

  it("Validate Signature", async function () {
    let gasUsedAverage = 0n;
    const amountRuns = 100;
    for (let i = 0; i < amountRuns; i++) {
      const [success, gasUsed] = await crypto.validateSignature(
        signedData[i].GroupSignature,
        signedData[i].BClaims
      );
      expect(success).to.be.equal(true);
      console.log(gasUsed);
      gasUsedAverage += gasUsed.toBigInt();
    }
    gasUsedAverage /= BigInt(amountRuns);
    console.log(gasUsedAverage);
  });

  it("Validate Signature ASM", async function () {
    let gasUsedAverage = 0n;
    const amountRuns = 100;
    for (let i = 0; i < amountRuns; i++) {
      const [success, gasUsed] = await crypto.validateSignatureASM(
        signedData[i].GroupSignature,
        signedData[i].BClaims
      );
      expect(success).to.be.equal(true);
      console.log(gasUsed);
      gasUsedAverage += gasUsed.toBigInt();
    }
    gasUsedAverage /= BigInt(amountRuns);
    console.log(gasUsedAverage);
  });

  //   it.only("Validate Signature ASM2", async function () {
  //     let gasUsedAverage = 0n;
  //     const amountRuns = 10;
  //     for (let i = 0; i < amountRuns; i++) {
  //       const [success, gasUsed] = await crypto.validateSignatureASM2(
  //         signedData[i].GroupSignature,
  //         signedData[i].BClaims
  //       );
  //       //expect(success).to.be.equal(true);
  //       console.log(gasUsed);
  //       gasUsedAverage += gasUsed.toBigInt();
  //     }
  //     gasUsedAverage /= BigInt(amountRuns);
  //     console.log(gasUsedAverage);
  //   });

  it("Validate Signature ASM3", async function () {
    let gasUsedAverage = 0n;
    const amountRuns = 100;
    for (let i = 0; i < amountRuns; i++) {
      const [success, gasUsed] = await crypto.validateSignatureASM3(
        signedData[i].GroupSignature,
        signedData[i].BClaims
      );
      expect(success).to.be.equal(true);
      console.log(gasUsed);
      gasUsedAverage += gasUsed.toBigInt();
    }
    gasUsedAverage /= BigInt(amountRuns);
    console.log(gasUsedAverage);
  });
});
